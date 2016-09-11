package gcloudcleanup

import (
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/urfave/cli.v2"
)

func TestNewCLI(t *testing.T) {
	c := NewCLI(&cli.Context{})
	assert.NotNil(t, c)
	assert.NotNil(t, c.c)
	assert.NotNil(t, c.log)
	assert.Nil(t, c.cs)
	assert.Nil(t, c.rateLimiter)
	assert.Nil(t, c.instanceCleaner)
	assert.Nil(t, c.imageCleaner)
}

func TestNewCLI_setupLogger(t *testing.T) {
	ranIt := false
	app := &cli.App{
		Action: func(c *cli.Context) error {
			gcccli := NewCLI(c)
			gcccli.log.Level = logrus.FatalLevel
			gcccli.setupLogger()
			assert.Equal(t, logrus.FatalLevel, gcccli.log.Level)
			ranIt = true
			return nil
		},
	}
	app.Run([]string{"foo"})
	assert.True(t, ranIt)
}

func TestNewCLI_setupRateLimiter_null(t *testing.T) {
	ranIt := false
	app := &cli.App{
		Flags: Flags,
		Action: func(c *cli.Context) error {
			gcccli := NewCLI(c)
			assert.Equal(t, "", c.String("rate-limit-redis-url"))
			gcccli.setupRateLimiter()
			tp := reflect.TypeOf(gcccli.rateLimiter)
			assert.Equal(t, "nullRateLimiter", tp.Name())
			ranIt = true
			return nil
		},
	}
	app.Run([]string{"foo"})
	assert.True(t, ranIt)
}

func TestNewCLI_setupRateLimiter_redis(t *testing.T) {
	ranIt := false
	app := &cli.App{
		Flags: Flags,
		Action: func(c *cli.Context) error {
			gcccli := NewCLI(c)
			assert.Equal(t, "redis://x:y@z.example.com:6379", c.String("rate-limit-redis-url"))
			gcccli.setupRateLimiter()
			assert.NotNil(t, gcccli.rateLimiter)
			tp := reflect.TypeOf(gcccli.rateLimiter)
			assert.NotEqual(t, "nullRateLimiter", tp.Name())
			ranIt = true
			return nil
		},
	}
	app.Run([]string{"foo", "--rate-limit-redis-url", "redis://x:y@z.example.com:6379"})
	assert.True(t, ranIt)
}
