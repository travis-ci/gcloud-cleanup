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

		maxAge := c.c.Duration("instance-max-age")
		cutoff := time.Now().UTC().Add(-1 * maxAge)

		if time.Now().UTC().Before(cutoff) {
			c.log.WithFields(logrus.Fields{
				"cutoff":  cutoff,
				"max_age": maxAge,
			}).Error("invalid instance max age given")
			return errInvalidInstancesMaxAge
		}

		c.log.WithFields(logrus.Fields{
			"max_age":    maxAge,
			"tick":       c.c.Duration("rate-tick-limit"),
			"project_id": c.c.String("project-id"),
			"filters":    strings.Join(filters, ","),
			"cutoff":     cutoff.Format(time.RFC3339),
		}).Debug("creating instance cleaner with")

		c.instanceCleaner = newInstanceCleaner(c.cs,
			c.log, c.c.Duration("rate-limit-tick"),
			cutoff, c.c.String("project-id"), filters)
	}

	return c.instanceCleaner.Run()
}

func (c *CLI) cleanupImages() error {
	if c.imageCleaner == nil {
		c.imageCleaner = newImageCleaner(c.cs, c.c.Duration("rate-limit-tick"))
	}
	return c.imageCleaner.Run()
}
