package gcloudcleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/urfave/cli.v2"
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
