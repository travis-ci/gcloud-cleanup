package gcloudcleanup

import (
	"time"

	"google.golang.org/api/compute/v1"
)

type imageCleaner struct {
	cs *compute.Service

	rateLimiter *time.Ticker
}

func newImageCleaner(cs *compute.Service, rlTick time.Duration) *imageCleaner {
	return &imageCleaner{
		cs: cs,

		rateLimiter: time.NewTicker(rlTick),
	}
}

func (ic *imageCleaner) Run() error {
	return nil
}

func (ic *imageCleaner) apiRateLimit() {
	<-ic.rateLimiter.C
}
