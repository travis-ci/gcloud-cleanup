package gcloudcleanup

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/gcloud-cleanup/ratelimit"
)

func TestNewInstanceCleaner(t *testing.T) {
	log := logrus.New()
	rl := ratelimit.NewNullRateLimiter()
	cutoffTime := time.Now().Add(-1 * time.Hour)

	ic := &instanceCleaner{
		log:               log.WithField("test", "yep"),
		rateLimiter:       rl,
		rateLimitMaxCalls: 10,
		rateLimitDuration: time.Second,
		CutoffTime:        cutoffTime,
		projectID:         "foo-project",
		filters:           []string{"name eq ^test.*"},
		noop:              true,
		archiveSerial:     true,
		archiveBucket:     "walrus-meme",
	}

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
	mux.HandleFunc(
		"/foo-project/zones/us-central1-a/instances/test-vm-1/serialPort",
		func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, "GET")
			fmt.Fprintf(w, `{}`)
		})
	mux.HandleFunc(
		"/foo-project/zones/us-central1-a/instances/test-vm-2/serialPort",
		func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, "GET")
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

	ctx := context.Background()
	sc, err := storage.NewClient(
		ctx, option.WithHTTPClient(&http.Client{Transport: &fakeTransport{}}))
	assert.Nil(t, err)

	log := logrus.New()
	log.Level = logrus.FatalLevel
	if os.Getenv("GCLOUD_CLEANUP_TEST_DEBUG") != "" {
		log.Level = logrus.DebugLevel
	}
	rl := ratelimit.NewNullRateLimiter()
	cutoffTime := time.Now().Add(-1 * time.Hour)

	ic := &instanceCleaner{
		cs:                cs,
		sc:                sc,
		log:               log.WithField("test", "yep"),
		rateLimiter:       rl,
		rateLimitMaxCalls: 10,
		rateLimitDuration: time.Second,
		CutoffTime:        cutoffTime,
		projectID:         "foo-project",
		filters:           []string{"name eq ^test.*"},
		noop:              false,
		archiveSerial:     true,
		archiveBucket:     "walrus-meme",
	}

	err = ic.Run()
	assert.Nil(t, err)
}

// {
// lifted from:
// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/75763d24f38012ba2bb6f3966a39a6f0759a353c/storage/writer_test.go#L37-L68
type fakeTransport struct {
	gotReq  *http.Request
	gotBody []byte
	results []transportResult
}

type transportResult struct {
	res *http.Response
	err error
}

func (t *fakeTransport) addResult(res *http.Response, err error) {
	t.results = append(t.results, transportResult{res, err})
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.gotReq = req
	t.gotBody = nil
	if req.Body != nil {
		bytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		t.gotBody = bytes
	}
	if len(t.results) == 0 {
		return nil, fmt.Errorf("error handling request")
	}
	result := t.results[0]
	t.results = t.results[1:]
	return result.res, result.err
}

// }
