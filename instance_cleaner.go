package gcloudcleanup

import (
	"time"

	"google.golang.org/api/compute/v1"
)

type instanceCleaner struct {
	cs *compute.Service

	projectID, zoneName string
	rateLimiter         *time.Ticker
}

func newInstanceCleaner(cs *compute.Service, rlTick time.Duration, projectID, zoneName string) *instanceCleaner {
	return &instanceCleaner{
		cs: cs,

		projectID:   projectID,
		zoneName:    zoneName,
		rateLimiter: time.NewTicker(rlTick),
	}
}

func (ic *instanceCleaner) Run() error {
	instances, err := ic.fetchTerminatedInstances()
	if err != nil {
		return err
	}

	for _, inst := range instances {
		ic.apiRateLimit()
		err = ic.deleteInstance(inst)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ic *instanceCleaner) fetchTerminatedInstances() ([]string, error) {
	return []string{}, nil
}

func (ic *instanceCleaner) deleteInstance(inst string) error {
	_, err := ic.cs.Instances.Delete(ic.projectID, ic.zoneName, inst).Do()
	return err
}

func (ic *instanceCleaner) apiRateLimit() {
	<-ic.rateLimiter.C
}
