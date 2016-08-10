package main

import (
	"os"

	"github.com/travis-ci/gcloud-cleanup"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = gcloudcleanup.VersionString
	app.Flags = gcloudcleanup.Flags
	app.Action = func(c *cli.Context) {
		gcloudcleanup.NewCLI(c).Run()
	}
	app.Run(os.Args)
}
