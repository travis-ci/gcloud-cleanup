package gcloudcleanup

import (
	"time"

	"google.golang.org/api/compute/v1"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

type CLI struct {
	c   *cli.Context
	cs  *compute.Service
	log *logrus.Logger
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
	zones := c.c.StringSlice("zones")
	entities := c.c.StringSlice("entities")

	entityMap := map[string]func(string) error{
		"instances": c.cleanupInstances,
		"images":    c.cleanupImages,
	}

	for {
		for _, zone := range zones {
			for _, entity := range entities {
				if f, ok := entityMap[entity]; ok {
					c.log.WithFields(logrus.Fields{
						"type": entity,
						"zone": zone,
					}).Debug("entering entity loop")

					err := f(zone)

					if err != nil {
						c.log.WithFields(logrus.Fields{
							"type": entity,
							"zone": zone,
							"err":  err,
						}).Fatal("failure during entity cleanup")
					}
				} else {
					c.log.WithFields(logrus.Fields{
						"type": entity,
						"zone": zone,
					}).Fatal("unknown entity type")
				}
			}

			if once {
				break
			}
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

func (c *CLI) cleanupInstances(zone string) error {
	ic := newInstanceCleaner(c.cs, c.c.Duration("rate-limit-tick"),
		c.c.String("project-id"), zone)
	return ic.Run()
}

func (c *CLI) cleanupImages(zone string) error {
	ic := newImageCleaner(c.cs, c.c.Duration("rate-limit-tick"))
	return ic.Run()
}
