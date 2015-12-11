package gcloudcleanup

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

type CLI struct {
	c   *cli.Context
	log *logrus.Logger
}

func NewCLI(c *cli.Context) *CLI {
	log := logrus.New()
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{DisableColors: true}

	return &CLI{c: c, log: log}
}

func (c *CLI) Run() {
	if c.c.Bool("debug") {
		c.log.Level = logrus.DebugLevel
	}

	sleepDur := c.c.Duration("loop-sleep")

	entityMap := map[string]func() error{
		"instances": c.cleanupInstances,
		"images":    c.cleanupImages,
	}

	for {
		for _, entity := range c.c.StringSlice("entities") {
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
		}

		if c.c.Bool("once") {
			break
		}

		c.log.WithField("duration", sleepDur).Info("sleeping")
		time.Sleep(sleepDur)
	}
}

func (c *CLI) cleanupInstances() error {
	return nil
}

func (c *CLI) cleanupImages() error {
	return nil
}
