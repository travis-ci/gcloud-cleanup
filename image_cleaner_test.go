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

func TestNewImageCleaner(t *testing.T) {
	log := logrus.New()
	ratelimit := ratelimit.NewNullRateLimiter()

	ic := newImageCleaner(nil, log, ratelimit, 10, time.Second,
		"foo-project", "http://foo.example.com",
		[]string{"name eq ^travis-test.*"}, true)

	assert.NotNil(t, ic)
	assert.Nil(t, ic.cs)
	assert.NotNil(t, ic.log)
	assert.Equal(t, "foo-project", ic.projectID)
	assert.Equal(t, "http://foo.example.com", ic.jobBoardURL)
	assert.Equal(t, []string{"name eq ^travis-test.*"}, ic.filters)
}

func TestImageCleaner_Run(t *testing.T) {
	gceMux := http.NewServeMux()
	gceMux.HandleFunc(
		"/foo-project/global/images",
		func(w http.ResponseWriter, req *http.Request) {
			body := map[string]interface{}{
				"items": []interface{}{
					map[string]string{
						"name":   "travis-test-image-0",
						"status": "READY",
					},
					map[string]string{
						"name":   "travis-test-bananapants-9001",
						"status": "READY",
					},
					map[string]string{
						"name":   "travis-test-bananapants-9000",
						"status": "READY",
					},
				},
			}
			err := json.NewEncoder(w).Encode(body)
			assert.Nil(t, err)
		})
	gceMux.HandleFunc("/foo-project/global/images/travis-test-image-0",
		func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, "DELETE")
			fmt.Fprintf(w, `{}`)
		})
	gceMux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			t.Errorf("Unhandled gce URL: %s %v", req.Method, req.URL)
		})

	gceSrv := httptest.NewServer(gceMux)
	defer gceSrv.Close()

	jbMux := http.NewServeMux()
	jbMux.HandleFunc("/images", func(w http.ResponseWriter, req *http.Request) {
		body := map[string]interface{}{
			"data": []interface{}{
				map[string]string{
					"name": "travis-test-bananapants-9000",
				},
				map[string]string{
					"name": "travis-test-bananapants-9001",
				},
				map[string]string{
					"name": "travis-test-roboticshoe-1138",
				},
			},
		}
		err := json.NewEncoder(w).Encode(body)
		assert.Nil(t, err)
	})

	jbMux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			t.Errorf("Unhandled job-board URL: %s %v", req.Method, req.URL)
		})

	jbSrv := httptest.NewServer(jbMux)
	defer jbSrv.Close()

	cs, err := compute.New(&http.Client{})
	assert.Nil(t, err)
	cs.BasePath = gceSrv.URL

	log := logrus.New()
	log.Level = logrus.FatalLevel
	if os.Getenv("GCLOUD_CLEANUP_TEST_DEBUG") != "" {
		log.Level = logrus.DebugLevel
	}
	rl := ratelimit.NewNullRateLimiter()

	ic := newImageCleaner(cs, log, rl, 10, time.Second,
		"foo-project", jbSrv.URL,
		[]string{"name eq ^travis-test.*"}, false)

	err = ic.Run()
	assert.Nil(t, err)
}
