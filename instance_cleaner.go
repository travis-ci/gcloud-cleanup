package gcloudcleanup

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v1"

	"github.com/sirupsen/logrus"
	"github.com/travis-ci/gcloud-cleanup/metrics"
	"github.com/travis-ci/gcloud-cleanup/ratelimit"
)

var (
	errNoStorageClient = fmt.Errorf("no storage client available")
)

type instanceCleaner struct {
	ctx context.Context
	cs  *compute.Service
	sc  *storage.Client
	log *logrus.Entry

	rand *rand.Rand

	projectID string
	filters   []string

	noop bool

	archiveSerial     bool
	archiveBucket     string
	archiveSampleRate int64

	CutoffTime time.Time

	rateLimiter       ratelimit.RateLimiter
	rateLimitMaxCalls uint64
	rateLimitDuration time.Duration
}

type instanceDeletionRequest struct {
	Instance *compute.Instance
	Reason   string
}

func (ic *instanceCleaner) Run() error {

	ctx, span := trace.StartSpan(context.Background(), "InstanceCleanerRun")
	defer span.End()

	ic.log.WithFields(logrus.Fields{
		"project":     ic.projectID,
		"cutoff_time": ic.CutoffTime.Format(time.RFC3339),
		"filters":     strings.Join(ic.filters, ","),
	}).Info("running instance cleanup")

	instChan := make(chan *instanceDeletionRequest)
	errChan := make(chan error)

	go ic.fetchInstancesToDelete(ctx, instChan, errChan)
	go func() {

		for err := range errChan {
			ic.log.WithField("err", err).Warn("error during instance fetch")
		}
	}()

	nDeleted := 0

	for req := range instChan {
		err := ic.deleteInstance(ctx, req.Instance)

		if err != nil {
			ic.log.WithFields(logrus.Fields{
				"err":      err,
				"instance": req.Instance.Name,
			}).Warn("failed to delete instance")
			continue
		}

		nDeleted++

		ic.log.WithFields(logrus.Fields{
			"instance": req.Instance.Name,
			"reason":   req.Reason,
		}).Info("deleted")
	}

	metrics.Counter("travis.gcloud-cleanup.instances.deleted", int64(nDeleted))
	ic.l2met("measure#instances.deleted", nDeleted, "done running instance cleanup")

	return nil
}

func (ic *instanceCleaner) fetchInstancesToDelete(ctx context.Context, instChan chan *instanceDeletionRequest, errChan chan error) {

	ctx, span := trace.StartSpan(ctx, "FetchInstancesToDelete")
	defer span.End()

	defer close(errChan)
	defer close(instChan)

	listCall := ic.cs.Instances.AggregatedList(ic.projectID)
	for _, filter := range ic.filters {
		listCall.Filter(filter)
	}

	pageTok := ""
	statusCounts := map[string]int{}
	nInstances := 0

	for {
		if pageTok != "" {
			listCall.PageToken(pageTok)
		}

		ic.apiRateLimit()
		ic.log.WithField("page_token", pageTok).Debug("fetching instances aggregated list")
		resp, err := listCall.Do()

		if err != nil {
			errChan <- err
			continue
		}

		ic.log.WithField("zones", len(resp.Items)).Debug("checking aggregated instance results")

		for zone, list := range resp.Items {
			ic.log.WithFields(logrus.Fields{
				"zone":      zone,
				"instances": len(list.Instances),
			}).Debug("checking instance results in zone")

			for _, inst := range list.Instances {
				nInstances++

				log := ic.log.WithFields(logrus.Fields{
					"instance": inst.Name,
				})

				if _, ok := statusCounts[inst.Status]; !ok {
					statusCounts[inst.Status] = 0
				}

				statusCounts[inst.Status]++

				if inst.Status == "TERMINATED" {
					log.WithFields(logrus.Fields{
						"status": inst.Status,
					}).Debug("sending instance for deletion")

					instChan <- &instanceDeletionRequest{Instance: inst, Reason: "terminated"}
					continue
				}

				ts, err := time.Parse(time.RFC3339, inst.CreationTimestamp)

				if err != nil {
					log.WithField("err", err).Warn("failed to parse creation timestamp")
					continue
				}

				ts = ts.UTC()

				log.WithFields(logrus.Fields{
					"orig":   inst.CreationTimestamp,
					"parsed": ts.Format(time.RFC3339),
				}).Debug("parsed and adjusted creation timestamp")

				if ts.Before(ic.CutoffTime) {
					log.WithFields(logrus.Fields{
						"created": ts.Format(time.RFC3339),
						"cutoff":  ic.CutoffTime.Format(time.RFC3339),
					}).Debug("sending instance for deletion")

					instChan <- &instanceDeletionRequest{Instance: inst, Reason: "stale"}
					continue
				}

				log.Debug("skipping instance")
			}
		}

		if resp.NextPageToken == "" {
			ic.log.Debug("no next page, breaking out of loop")
			break
		}

		ic.log.Debug("continuing to next page")
		pageTok = resp.NextPageToken
	}

	for status, count := range statusCounts {
		key := fmt.Sprintf("gauge#instances.status.%s", status)
		ic.l2met(key, count, "counted instances with status")
	}

	ic.l2met("gauge#instances.count", nInstances, "done checking all instances")
}

func (ic *instanceCleaner) deleteInstance(ctx context.Context, inst *compute.Instance) error {
	// instance cleaner run on the outside, delete
	ctx, span := trace.StartSpan(ctx, "DeleteInstance")
	defer span.End()
	if ic.noop {
		ic.log.WithField("instance", inst.Name).Debug("not really deleting instance")
		return nil
	}

	if ic.archiveSerial {
		ic.log.WithField("instance", inst.Name).Debug("archiving serial port output")
		err := ic.archiveSerialConsoleOutput(inst)
		if err != nil {
			return err
		}
	}

	ic.apiRateLimit()
	_, err := ic.cs.Instances.Delete(ic.projectID, filepath.Base(inst.Zone), inst.Name).Do()
	return err
}

func (ic *instanceCleaner) l2met(name string, n int, msg string) {
	ic.log.WithField(name, n).Info(msg)
}

func (ic *instanceCleaner) archiveSerialConsoleOutput(inst *compute.Instance) error {
	if ic.sc == nil {
		return errNoStorageClient
	}

	archiveSampled := ic.rand.Float32() < (1.0 / float32(ic.archiveSampleRate))

	if !archiveSampled {
		ic.log.WithField("instance", inst.Name).Debug("skipping archive due to sample rate")
		return nil
	}

	accum := ""
	lastPos := int64(0)

	for {
		ic.apiRateLimit()
		resp, err := ic.cs.Instances.GetSerialPortOutput(
			ic.projectID, filepath.Base(inst.Zone), inst.Name).Start(lastPos).Context(ic.ctx).Do()

		if err != nil {
			return err
		}

		accum += resp.Contents
		if lastPos == resp.Next {
			break
		}
		lastPos = resp.Next
	}

	key := fmt.Sprintf("serial-console-output/%s.txt", inst.Name)
	obj := ic.sc.Bucket(ic.archiveBucket).Object(key)
	wc := obj.NewWriter(ic.ctx)

	_, err := io.Copy(wc, strings.NewReader(accum))
	if err != nil {
		ic.log.WithFields(logrus.Fields{
			"err":      err,
			"instance": inst.Name,
		}).Warn("failed to copy console output to archive")
		return err
	}

	err = wc.Close()
	if err != nil {
		ic.log.WithFields(logrus.Fields{
			"err":      err,
			"instance": inst.Name,
		}).Warn("failed to close console output upload writer")
		return err
	}

	return nil
}

func (ic *instanceCleaner) apiRateLimit() error {
	ic.log.Debug("waiting for rate limiter tick")
	errCount := 0

	for {
		ok, err := ic.rateLimiter.RateLimit("gce-api", ic.rateLimitMaxCalls, ic.rateLimitDuration)
		if err != nil {
			errCount++
			if errCount >= 5 {
				ic.log.WithField("err", err).Info("rate limiter errored 5 times")
				return err
			}
		} else {
			errCount = 0
		}
		if ok {
			return nil
		}

		// Sleep for up to 1 second
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	}
}
