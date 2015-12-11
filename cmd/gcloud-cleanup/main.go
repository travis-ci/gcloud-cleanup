package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/travis-ci/gcloud-cleanup"
)

func main() {
	app := cli.NewApp()
	app.Flags = gcloudcleanup.Flags
	app.Action = func(c *cli.Context) {
		gcloudcleanup.NewCLI(c).Run()
	}
	app.Run(os.Args)
}
