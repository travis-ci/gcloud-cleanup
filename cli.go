package gcloudcleanup

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	errInvalidInstancesMaxAge = fmt.Errorf("invalid max age")
)

type CLI struct {
	c   *cli.Context
	cs  *compute.Service
	log *logrus.Logger

	instanceCleaner *instanceCleaner
	imageCleaner    *imageCleaner
}

func NewCLI(c *cli.Context) *CLI {
	log := logrus.New()
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{DisableColors: true}

	return &CLI{c: c, log: log}
}

func (c *CLI) Run() {
	err := c.setupComputeService(c.c.String("account-json"))
	if err != nil {
		c.log.WithField("err", err).Fatal("failed to set up compute service")
	}

	c.setupLogger()

	sleepDur := c.c.Duration("loop-sleep")
	if sleepDur == (0 * time.Second) {
		sleepDur = 5 * time.Minute
		c.log.WithField("loop_sleep", sleepDur).Info("default loop sleep set")
	}

	once := c.c.Bool("once")

	entities := c.c.StringSlice("entities")
	if len(entities) == 0 {
		entities = []string{"instances", "images"}
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
}

func (c *CLI) setupComputeService(accountJSON string) error {
	cs, err := buildGoogleComputeService(accountJSON)
	c.cs = cs
	return err
}

func (c *CLI) setupLogger() {
	if c.c.Bool("debug") {
		c.log.Level = logrus.DebugLevel
	}
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

		c.log.WithFields(logrus.Fields{
			"max_age":    c.c.Duration("instance-max-age"),
			"tick":       c.c.Duration("rate-tick-limit"),
			"project_id": c.c.String("project-id"),
			"filters":    strings.Join(filters, ","),
			"cutoff":     cutoffTime.Format(time.RFC3339),
		}).Debug("creating instance cleaner with")

		c.instanceCleaner = newInstanceCleaner(c.cs,
			c.log, c.c.Duration("rate-limit-tick"),
			cutoffTime, c.c.String("project-id"), filters)
	}

	c.instanceCleaner.CutoffTime = time.Now().UTC().Add(-1 * c.c.Duration("instance-max-age"))

	return c.instanceCleaner.Run()
}

func (c *CLI) cleanupImages() error {
	if c.imageCleaner == nil {
		filters := c.c.StringSlice("image-filters")
		if len(filters) == 0 {
			filters = []string{"name eq ^travis-ci.*"}
			c.log.WithField("filters", strings.Join(filters, ",")).Info("default filters set")
		}

		c.imageCleaner = newImageCleaner(c.cs,
			c.log, c.c.Duration("rate-limit-tick"), c.c.String("project-id"),
			c.c.String("job-board-url"), filters)
	}

	return c.imageCleaner.Run()
}
