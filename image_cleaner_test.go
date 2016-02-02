package gcloudcleanup

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/ratelimit"
)

func TestNewImageCleaner(t *testing.T) {
	log := logrus.New()
	ratelimit := ratelimit.NewNullRateLimiter()

	ic := newImageCleaner(nil, log, ratelimit, 10, time.Second,
		"foo-project", "http://foo.example.com", 20,
		[]string{"name eq ^travis-test.*"}, true)

	assert.NotNil(t, ic)
	assert.Nil(t, ic.cs)
	assert.NotNil(t, ic.log)
	assert.Equal(t, "foo-project", ic.projectID)
	assert.Equal(t, "http://foo.example.com", ic.jobBoardURL)
	assert.Equal(t, 20, ic.imageLimit)
	assert.Equal(t, []string{"name eq ^travis-test.*"}, ic.filters)
}
