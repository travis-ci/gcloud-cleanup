package gcloudcleanup

import (
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/ratelimit"
)

func TestNewInstanceCleaner(t *testing.T) {
	log := logrus.New()
	ratelimit := ratelimit.NewNullRateLimiter()
	cutoffTime := time.Now().Add(-1 * time.Hour)

	ic := newInstanceCleaner(nil, log, ratelimit, 10, time.Second,
		cutoffTime, "foo-project",
		[]string{"name eq ^test.*"}, true)

	assert.NotNil(t, ic)
	assert.NotNil(t, ic.log)
	assert.Equal(t, "foo-project", ic.projectID)
	assert.Equal(t, []string{"name eq ^test.*"}, ic.filters)
	assert.True(t, ic.noop)
	assert.Equal(t, cutoffTime, ic.CutoffTime)
}
