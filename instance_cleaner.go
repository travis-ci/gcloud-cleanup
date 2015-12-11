package gcloudcleanup

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"google.golang.org/api/compute/v1"
)

type instanceCleaner struct {
	cs  *compute.Service
	log *logrus.Logger

	projectID   string
	rateLimiter *time.Ticker
	cutoffTime  time.Time
	filters     []string
}

func newInstanceCleaner(cs *compute.Service, log *logrus.Logger,
	rlTick time.Duration, cutoffTime time.Time,
	projectID string, filters []string) *instanceCleaner {

	return &instanceCleaner{
		cs:  cs,
		log: log,

		projectID:   projectID,
		rateLimiter: time.NewTicker(rlTick),
		cutoffTime:  cutoffTime,
		filters:     filters,
	}
}

func (ic *instanceCleaner) Run() error {
	ic.log.WithFields(logrus.Fields{
		"project":     ic.projectID,
		"cutoff_time": ic.cutoffTime.Format(time.RFC3339),
		"filters":     strings.Join(ic.filters, ","),
	}).Info("running instance cleanup")

	instChan := make(chan *compute.Instance)
	errChan := make(chan error)

	go ic.fetchInstancesToDelete(instChan, errChan)
	go func() {
		for err := range errChan {
			if err == nil {
				continue
			}
			ic.log.WithField("err", err).Warn("error during instance fetch")
		}
	}()

	for inst := range instChan {
		if inst == nil {
			return nil
		}

		err := ic.deleteInstance(inst)

		if err != nil {
			ic.log.WithFields(logrus.Fields{
				"err":      err,
				"instance": inst.Name,
			}).Warn("failed to delete instance")
		}

		ic.log.WithField("instance", inst.Name).Info("deleted")
	}

	return nil
}

func (ic *instanceCleaner) fetchInstancesToDelete(instChan chan *compute.Instance, errChan chan error) {
	listCall := ic.cs.Instances.AggregatedList(ic.projectID)
	for _, filter := range ic.filters {
		listCall.Filter(filter)
	}

	pageTok := ""

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
				log := ic.log.WithFields(logrus.Fields{
					"instance":  inst.Name,
					"component": "instance_cleaner",
				})

				if inst.Status == "TERMINATED" {
					log.WithFields(logrus.Fields{
						"status": inst.Status,
					}).Debug("sending instance for deletion")

					instChan <- inst
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

				if ts.Before(ic.cutoffTime) {
					log.WithFields(logrus.Fields{
						"created": ts.Format(time.RFC3339),
						"cutoff":  ic.cutoffTime.Format(time.RFC3339),
					}).Debug("sending instance for deletion")

					instChan <- inst
					continue
				}

				log.Debug("skipping instance")
			}
		}

		if resp.NextPageToken == "" {
			ic.log.Debug("no next page, breaking out of loop")
			instChan <- nil
			errChan <- nil
			return
		}

		ic.log.Debug("continuing to next page")
		pageTok = resp.NextPageToken
	}
}

func (ic *instanceCleaner) deleteInstance(inst *compute.Instance) error {
	ic.apiRateLimit()

	_, err := ic.cs.Instances.Delete(ic.projectID, filepath.Base(inst.Zone), inst.Name).Do()
	return err
}

func (ic *instanceCleaner) apiRateLimit() {
	ic.log.Debug("waiting for rate limiter tick")
	<-ic.rateLimiter.C
}
