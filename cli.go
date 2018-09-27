package gcloudcleanup

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/urfave/cli.v2"

	"github.com/mihasya/go-metrics-librato"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	travismetrics "github.com/travis-ci/gcloud-cleanup/metrics"
	"github.com/travis-ci/gcloud-cleanup/ratelimit"
)

var (
	errInvalidInstancesMaxAge   = errors.New("invalid max age")
	errInvalidArchiveSampleRate = errors.New("invalid archive sample rate")
	errInvalidTraceSampleRate   = errors.New("invalid trace sample rate")
)

type CLI struct {
	c           *cli.Context
	ctx         context.Context
	cs          *compute.Service
	sc          *storage.Client
	log         *logrus.Logger
	rateLimiter ratelimit.RateLimiter

	instanceCleaner *instanceCleaner
	imageCleaner    *imageCleaner
}

func NewCLI(c *cli.Context) *CLI {
	log := logrus.New()
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{DisableColors: true}

	return &CLI{
		c:   c,
		ctx: context.Background(),
		log: log,
	}
}

func (c *CLI) Run() error {
	c.setupLogger()
	c.setupRateLimiter()

	fields := logrus.Fields{}

	for _, name := range c.c.FlagNames() {
		fields[name] = c.c.Generic(name)
	}

	c.log.WithFields(fields).Debug("configuration")

	c.setupMetrics()

	err := c.setupOpenCensus(c.c.String("account-json"))
	if err != nil {
		c.log.WithField("err", err).Fatal("failed to set up opencensus")
	}

	err = c.setupComputeService(c.c.String("account-json"))
	if err != nil {
		c.log.WithField("err", err).Fatal("failed to set up compute service")
	}

	err = c.setupStorageClient(c.c.String("account-json"))
	if err != nil {
		c.log.WithField("err", err).Fatal("failed to set up storage client")
	}

	sleepDur := c.c.Duration("loop-sleep")
	if sleepDur == (0 * time.Second) {
		sleepDur = 5 * time.Minute
		c.log.WithField("loop_sleep", sleepDur).Info("default loop sleep set")
	}

	once := c.c.Bool("once")

	entities := c.c.StringSlice("entities")
	if len(entities) == 0 {
		entities = []string{"instances"}
		c.log.WithField("entities", entities).Info("default entities set")
	}

	entityMap := map[string]func() error{
		"instances": c.cleanupInstances,
		"images":    c.cleanupImages,
	}

	for {
		for _, entity := range entities {
			if f, ok := entityMap[entity]; ok {
				c.log.WithField("type", entity).Debug("entering entity loop")

				err := f()

				if err != nil {
					c.log.WithFields(logrus.Fields{
						"type": entity,
						"err":  err,
					}).Fatal("failure during entity cleanup")
				}
			} else {
				c.log.WithField("type", entity).Fatal("unknown entity type")
			}

			c.log.WithField("type", entity).Debug("done with entity loop")
		}

		if once {
			break
		}

		c.log.WithField("duration", sleepDur).Info("sleeping")
		time.Sleep(sleepDur)
	}
	return nil
}

func (c *CLI) setupComputeService(accountJSON string) error {
	cs, err := buildGoogleComputeService(accountJSON)
	c.cs = cs
	return err
}

func (c *CLI) setupStorageClient(accountJSON string) error {
	sc, err := buildGoogleStorageClient(c.ctx, accountJSON)
	c.sc = sc
	return err
}

func (c *CLI) setupLogger() {
	if c.c.Bool("debug") {
		c.log.Level = logrus.DebugLevel
	}
}

func (c *CLI) setupRateLimiter() {
	if c.c.String("rate-limit-redis-url") == "" {
		c.rateLimiter = ratelimit.NewNullRateLimiter()
		return
	}
	c.rateLimiter = ratelimit.NewRateLimiter(
		c.c.String("rate-limit-redis-url"),
		c.c.String("rate-limit-prefix"))
}

func (c *CLI) setupOpenCensus(accountJSON string) error {
	opencensusEnabled := c.c.Bool("opencensus-tracing-enabled")

	if !opencensusEnabled {
		return nil
	}

	creds, err := buildGoogleCloudCredentials(context.TODO(), accountJSON)
	if err != nil {
		return err
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GCLOUD_PROJECT"),
		TraceClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
		MonitoringClientOptions: []option.ClientOption{
			option.WithCredentials(creds),
		},
	})

	if err != nil {
		return err
	}

	defer sd.Flush()

	// Register/enable the trace exporter
	trace.RegisterExporter(sd)

	traceSampleRate := c.c.Int64("opencensus-sampling-rate")
	if traceSampleRate <= 0 {
		c.log.WithFields(logrus.Fields{
			"trace_sample_rate": traceSampleRate,
		}).Error("trace sample rate must be positive")
		return errInvalidTraceSampleRate
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1.0 / float64(traceSampleRate))})

	return nil
}

func (c *CLI) cleanupInstances() error {
	if c.instanceCleaner == nil {
		filters := c.c.StringSlice("instance-filters")
		if len(filters) == 0 {
			filters = []string{"name eq ^testing-gce.*"}
			c.log.WithField("filters", strings.Join(filters, ",")).Info("default filters set")
		}

		cutoffTime := time.Now().UTC().Add(-1 * c.c.Duration("instance-max-age"))

		if time.Now().UTC().Before(cutoffTime) {
			c.log.WithFields(logrus.Fields{
				"cutoff":  cutoffTime,
				"max_age": c.c.Duration("instance-max-age"),
			}).Error("invalid instance max age given")
			return errInvalidInstancesMaxAge
		}

		archiveSampleRate := c.c.Int64("archive-sample-rate")
		if archiveSampleRate <= 0 {
			c.log.WithFields(logrus.Fields{
				"sample_rate": archiveSampleRate,
			}).Error("archive sample rate must be positive")
			return errInvalidArchiveSampleRate
		}

		c.log.WithFields(logrus.Fields{
			"max_age":    c.c.Duration("instance-max-age"),
			"tick":       c.c.Duration("rate-tick-limit"),
			"project_id": c.c.String("project-id"),
			"filters":    strings.Join(filters, ","),
			"cutoff":     cutoffTime.Format(time.RFC3339),
		}).Debug("creating instance cleaner with")

		c.instanceCleaner = &instanceCleaner{
			ctx: c.ctx,
			cs:  c.cs,
			sc:  c.sc,
			log: c.log.WithField("component", "instance_cleaner"),

			rand: rand.New(rand.NewSource(time.Now().UnixNano())),

			projectID: c.c.String("project-id"),
			filters:   filters,

			archiveSerial:     c.c.Bool("archive-serial"),
			archiveBucket:     c.c.String("archive-bucket"),
			archiveSampleRate: archiveSampleRate,
			noop:              c.c.Bool("noop"),

			CutoffTime: cutoffTime,

			rateLimiter:       c.rateLimiter,
			rateLimitMaxCalls: uint64(c.c.Int("rate-limit-max-calls")),
			rateLimitDuration: c.c.Duration("rate-limit-duration"),
		}
	}

	c.instanceCleaner.CutoffTime = time.Now().UTC().Add(-1 * c.c.Duration("instance-max-age"))

	return c.instanceCleaner.Run()
}

func (i *CLI) setupMetrics() {
	go travismetrics.ReportMemstatsMetrics()

	if os.Getenv("LIBRATO_EMAIL") != "" && os.Getenv("LIBRATO_TOKEN") != "" && os.Getenv("LIBRATO_SOURCE") != "" {
		i.log.Info("starting librato metrics reporter")

		go librato.Librato(metrics.DefaultRegistry, time.Minute,
			os.Getenv("LIBRATO_EMAIL"), os.Getenv("LIBRATO_TOKEN"), os.Getenv("LIBRATO_SOURCE"),
			[]float64{0.50, 0.75, 0.90, 0.95, 0.99, 0.999, 1.0}, time.Millisecond)
	}
}

func (c *CLI) cleanupImages() error {
	if c.imageCleaner == nil {
		filters := c.c.StringSlice("image-filters")
		if len(filters) == 0 {
			filters = []string{"name eq ^travis-ci.*"}
			c.log.WithField("filters", strings.Join(filters, ",")).Info("default filters set")
		}

		c.imageCleaner = newImageCleaner(c.cs,
			c.log, c.rateLimiter, uint64(c.c.Int("rate-limit-max-calls")), c.c.Duration("rate-limit-duration"), c.c.String("project-id"),
			c.c.String("job-board-url"), filters, c.c.Bool("noop"))
	}

	return c.imageCleaner.Run()
}
