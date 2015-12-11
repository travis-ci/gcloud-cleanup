package gcloudcleanup

import (
	"fmt"

	"github.com/codegangsta/cli"
)

type CLI struct {
	c      *cli.Context
	config map[string]string
}

func NewCLI(c *cli.Context) *CLI {
	return &CLI{c: c, config: map[string]string{}}
}

func (c *CLI) Run() {
	fmt.Println("ohai there")
}
