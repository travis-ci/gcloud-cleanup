package main

import (
	"os"

	"github.com/travis-ci/gcloud-cleanup"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Usage:     "Google Compute Platform cleaner thingy",
		Version:   gcloudcleanup.VersionString,
		Copyright: gcloudcleanup.CopyrightString,
		Flags:     gcloudcleanup.Flags,
		Action: func(c *cli.Context) error {
			return gcloudcleanup.NewCLI(c).Run()
		},
	}
	app.Run(os.Args)
}
