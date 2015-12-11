package gcloudcleanup

import (
	"time"

	"github.com/codegangsta/cli"
)

var (
	Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "zones",
			Value:  &cli.StringSlice{"us-central1-b", "us-central1-f"},
			Usage:  "zones to clean up",
			EnvVar: "GCLOUD_CLEANUP_ZONES",
		},
		cli.StringSliceFlag{
			Name:   "entities",
			Value:  &cli.StringSlice{"instances", "images"},
			Usage:  "entities to clean up",
			EnvVar: "GCLOUD_CLEANUP_ENTITIES",
		},
		cli.DurationFlag{
			Name:   "loop-sleep",
			Value:  10 * time.Second,
			Usage:  "sleep time between loops",
			EnvVar: "GCLOUD_CLEANUP_LOOP_SLEEP",
		},
		cli.BoolFlag{
			Name:   "once",
			Usage:  "only run once, no looping",
			EnvVar: "GCLOUD_CLEANUP_ONCE",
		},
		cli.IntFlag{
			Name:   "image-limit",
			Value:  100,
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
	}
)
