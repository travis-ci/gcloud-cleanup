package gcloudcleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestNewCLI(t *testing.T) {
	c := NewCLI(&cli.Context{})
	assert.NotNil(t, c)
	assert.NotNil(t, c.c)
	assert.NotNil(t, c.log)
	assert.Nil(t, c.cs)
	assert.Nil(t, c.instanceCleaner)
	assert.Nil(t, c.imageCleaner)
}
