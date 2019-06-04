package gcloudcleanup

import (
	"time"

	"gopkg.in/urfave/cli.v2"
)

var (
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "account-json",
			Value:   "",
			Usage:   "file path to or json blob of GCE account stuff",
			EnvVars: []string{"GCLOUD_CLEANUP_ACCOUNT_JSON", "GOOGLE_CREDENTIALS"},
		},
		&cli.StringFlag{
			Name:    "project-id",
			Value:   "",
			Usage:   "name of GCE project",
			EnvVars: []string{"GCLOUD_CLEANUP_PROJECT_ID", "GCLOUD_PROJECT"},
		},
		&cli.DurationFlag{
			Name:    "instance-max-age",
			Value:   3 * time.Hour,
			Usage:   "max age for an instance to be considered deletable",
			EnvVars: []string{"GCLOUD_CLEANUP_INSTANCE_MAX_AGE"},
		},
		&cli.StringSliceFlag{
			Name:    "instance-filters",
			Usage:   "filters used when fetching instances for deletion",
			EnvVars: []string{"GCLOUD_CLEANUP_INSTANCE_FILTERS"},
		},
		&cli.StringSliceFlag{
			Name:    "image-filters",
			Usage:   "filters used when fetching images for deletion",
			EnvVars: []string{"GCLOUD_CLEANUP_IMAGE_FILTERS"},
		},
		&cli.StringSliceFlag{
			Name:    "entities",
			Usage:   "entities to clean up",
			EnvVars: []string{"GCLOUD_CLEANUP_ENTITIES"},
		},
		&cli.DurationFlag{
			Name:    "loop-sleep",
			Value:   5 * time.Minute,
			Usage:   "sleep time between loops",
			EnvVars: []string{"GCLOUD_CLEANUP_LOOP_SLEEP"},
		},
		&cli.BoolFlag{
			Name:    "once",
			Usage:   "only run once, no looping",
			EnvVars: []string{"GCLOUD_CLEANUP_ONCE"},
		},
		&cli.StringFlag{
			Name:    "rate-limit-redis-url",
			Usage:   "URL to Redis instance to use for rate limiting",
			EnvVars: []string{"GCLOUD_CLEANUP_RATE_LIMIT_REDIS_URL"},
		},
		&cli.StringFlag{
			Name:    "rate-limit-prefix",
			Usage:   "prefix for the rate limit key in Redis",
			EnvVars: []string{"GCLOUD_CLEANUP_RATE_LIMIT_PREFIX"},
		},
		&cli.IntFlag{
			Name:    "rate-limit-max-calls",
			Value:   10,
			Usage:   "number of calls per duration to let through to the GCE API",
			EnvVars: []string{"GCLOUD_CLEANUP_RATE_LIMIT_MAX_CALLS"},
		},
		&cli.DurationFlag{
			Name:    "rate-limit-duration",
			Value:   1 * time.Second,
			Usage:   "interval in which to let max-calls through to the GCE API",
			EnvVars: []string{"GCLOUD_CLEANUP_RATE_LIMIT_DURATION"},
		},
		&cli.StringFlag{
			Name:    "job-board-url",
			Value:   "http://localhost:4567",
			Usage:   "url to job-board instance for fetching registered images",
			EnvVars: []string{"GCLOUD_CLEANUP_JOB_BOARD_URL", "JOB_BOARD_URL"},
		},
		&cli.BoolFlag{
			Name:    "archive-serial",
			Usage:   "archive instance serial output before deleting",
			EnvVars: []string{"GCLOUD_CLEANUP_ARCHIVE_SERIAL", "ARCHIVE_SERIAL"},
		},
		&cli.StringFlag{
			Name:    "archive-bucket",
			Value:   "gcloud-cleanup-serial-output",
			Usage:   "bucket to which instance serial output will be archived before deleting",
			EnvVars: []string{"GCLOUD_CLEANUP_ARCHIVE_BUCKET", "ARCHIVE_BUCKET"},
		},
		&cli.Int64Flag{
			Name:    "archive-sample-rate",
			Value:   1,
			Usage:   "sample rate for archiving as an inverse fraction - for sample rate n, every nth event will be sampled",
			EnvVars: []string{"GCLOUD_CLEANUP_ARCHIVE_SAMPLE_RATE", "ARCHIVE_SAMPLE_RATE"},
		},
		&cli.Int64Flag{
			Name:    "opencensus-sampling-rate",
			Value:   1,
			Usage:   "sample rate for trace as an inverse fraction - for sample rate n, every nth event will be sampled",
			EnvVars: []string{"GCLOUD_CLEANUP_OPENCENSUS_SAMPLING_RATE", "OPENCENSUS_SAMPLING_RATE"},
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "output more stuff",
			EnvVars: []string{"GCLOUD_CLEANUP_DEBUG", "DEBUG"},
		},
		&cli.BoolFlag{
			Name:    "opencensus-tracing-enabled",
			Usage:   "enable tracing for gcloud-cleanup",
			EnvVars: []string{"GCLOUD_CLEANUP_OPENCENSUS_TRACING_ENABLED", "OPENCENSUS_TRACING_ENABLED"},
		},
		&cli.BoolFlag{
			Name:    "noop",
			Usage:   "don't do mutative stuff",
			EnvVars: []string{"GCLOUD_CLEANUP_NOOP", "NOOP"},
		},
		&cli.StringFlag{
			Name:    "librato-email",
			Usage:   "librato account for collecting metrics",
			EnvVars: []string{"GCLOUD_CLEANUP_LIBRATO_EMAIL", "LIBRATO_EMAIL"},
		},
		&cli.StringFlag{
			Name:    "librato-token",
			Usage:   "librato token for collecting metrics",
			EnvVars: []string{"GCLOUD_CLEANUP_LIBRATO_TOKEN", "LIBRATO_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "librato-source",
			Usage:   "librato source for collecting metrics",
			EnvVars: []string{"GCLOUD_CLEANUP_LIBRATO_SOURCE", "LIBRATO_SOURCE"},
		},
	}
)
