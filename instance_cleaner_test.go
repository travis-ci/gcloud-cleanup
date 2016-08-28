package gcloudcleanup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/ratelimit"
)

func TestNewInstanceCleaner(t *testing.T) {
	log := logrus.New()
	rl := ratelimit.NewNullRateLimiter()
	cutoffTime := time.Now().Add(-1 * time.Hour)

	ic := newInstanceCleaner(nil, log, rl, 10, time.Second,
		cutoffTime, "foo-project",
		[]string{"name eq ^test.*"}, true)

	assert.NotNil(t, ic)
	assert.NotNil(t, ic.log)
	assert.Equal(t, "foo-project", ic.projectID)
	assert.Equal(t, []string{"name eq ^test.*"}, ic.filters)
	assert.True(t, ic.noop)
	assert.Equal(t, cutoffTime, ic.CutoffTime)
}

func TestInstanceCleaner_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/foo-project/aggregated/instances",
		func(w http.ResponseWriter, req *http.Request) {
			body := map[string]interface{}{
				"items": map[string]interface{}{
					"zones/us-central1-a": map[string]interface{}{
						"instances": []interface{}{
							map[string]string{
								"name":              "test-vm-0",
								"status":            "RUNNING",
								"creationTimestamp": time.Now().Format(time.RFC3339),
								"zone":              "zones/us-central1-a",
							},
							map[string]string{
								"name":              "test-vm-1",
								"status":            "TERMINATED",
								"creationTimestamp": "2016-01-02T07:11:12.999-07:00",
								"zone":              "zones/us-central1-a",
							},
							map[string]string{
								"name":              "test-vm-2",
								"status":            "RUNNING",
								"creationTimestamp": time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
								"zone":              "zones/us-central1-a",
							},
						},
					},
				},
			}
			err := json.NewEncoder(w).Encode(body)
			assert.Nil(t, err)
		})
	mux.HandleFunc(
		"/foo-project/zones/us-central1-a/instances/test-vm-1",
		func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, "DELETE")
			fmt.Fprintf(w, `{}`)
		})
	mux.HandleFunc(
		"/foo-project/zones/us-central1-a/instances/test-vm-2",
		func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, "DELETE")
			fmt.Fprintf(w, `{}`)
		})
	mux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			t.Errorf("Unhandled URL: %s %v", req.Method, req.URL)
		})

	srv := httptest.NewServer(mux)

	defer srv.Close()

	cs, err := compute.New(&http.Client{})
	assert.Nil(t, err)
	cs.BasePath = srv.URL

	log := logrus.New()
	log.Level = logrus.FatalLevel
	if os.Getenv("GCLOUD_CLEANUP_TEST_DEBUG") != "" {
		log.Level = logrus.DebugLevel
	}
	rl := ratelimit.NewNullRateLimiter()
	cutoffTime := time.Now().Add(-1 * time.Hour)

	ic := newInstanceCleaner(cs, log, rl, 10, time.Second,
		cutoffTime, "foo-project",
		[]string{"name eq ^test.*"}, false)

	err = ic.Run()
	assert.Nil(t, err)
}
