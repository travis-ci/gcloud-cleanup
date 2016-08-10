package gcloudcleanup

import (
	"time"

	"github.com/urfave/cli"
)

var (
	Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "account-json",
			Value:  "{}",
			Usage:  "file path to or json blob of GCE account stuff",
			EnvVar: "GCLOUD_CLEANUP_ACCOUNT_JSON,GOOGLE_CREDENTIALS",
		},
		cli.StringFlag{
			Name:   "project-id",
			Usage:  "name of GCE project",
			EnvVar: "GCLOUD_CLEANUP_PROJECT_ID,GCLOUD_PROJECT",
		},
		cli.DurationFlag{
			Name:   "instance-max-age",
			Value:  3 * time.Hour,
			Usage:  "max age for an instance to be considered deletable",
			EnvVar: "GCLOUD_CLEANUP_INSTANCE_MAX_AGE",
		},
		cli.StringSliceFlag{
			Name:   "instance-filters",
			Usage:  "filters used when fetching instances for deletion",
			EnvVar: "GCLOUD_CLEANUP_INSTANCE_FILTERS",
		},
		cli.StringSliceFlag{
			Name:   "image-filters",
			Usage:  "filters used when fetching images for deletion",
			EnvVar: "GCLOUD_CLEANUP_IMAGE_FILTERS",
		},
		cli.StringSliceFlag{
			Name:   "entities",
			Usage:  "entities to clean up",
			EnvVar: "GCLOUD_CLEANUP_ENTITIES",
		},
		cli.DurationFlag{
			Name:   "loop-sleep",
			Value:  5 * time.Minute,
			Usage:  "sleep time between loops",
			EnvVar: "GCLOUD_CLEANUP_LOOP_SLEEP",
		},
		cli.BoolFlag{
			Name:   "once",
			Usage:  "only run once, no looping",
			EnvVar: "GCLOUD_CLEANUP_ONCE",
		},
		cli.StringFlag{
			Name:   "rate-limit-redis-url",
			Usage:  "URL to Redis instance to use for rate limiting",
			EnvVar: "GCLOUD_CLEANUP_RATE_LIMIT_REDIS_URL",
		},
		cli.StringFlag{
			Name:   "rate-limit-prefix",
			Usage:  "prefix for the rate limit key in Redis",
			EnvVar: "GCLOUD_CLEANUP_RATE_LIMIT_PREFIX",
		},
		cli.IntFlag{
			Name:   "rate-limit-max-calls",
			Value:  10,
			Usage:  "number of calls per duration to let through to the GCE API",
			EnvVar: "GCLOUD_CLEANUP_RATE_LIMIT_MAX_CALLS",
		},
		cli.DurationFlag{
			Name:   "rate-limit-duration",
			Value:  1 * time.Second,
			Usage:  "interval in which to let max-calls through to the GCE API",
			EnvVar: "GCLOUD_CLEANUP_RATE_LIMIT_DURATION",
		},
		cli.IntFlag{
			Name:   "image-limit",
			Value:  1000,
			Usage:  "number of images to fetch from job-board",
			EnvVar: "GCLOUD_CLEANUP_IMAGE_LIMIT",
		},
		cli.StringFlag{
			Name:   "job-board-url",
			Value:  "http://localhost:4567",
			Usage:  "url to job-board instance for fetching registered images",
			EnvVar: "GCLOUD_CLEANUP_JOB_BOARD_URL,JOB_BOARD_URL",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "output more stuff",
			EnvVar: "GCLOUD_CLEANUP_DEBUG,DEBUG",
		},
		cli.BoolFlag{
			Name:   "noop",
			Usage:  "don't do mutative stuff",
			EnvVar: "GCLOUD_CLEANUP_NOOP,NOOP",
		},
	}
)
