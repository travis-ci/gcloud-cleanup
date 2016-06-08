# gcloud-cleanup

Clean That Cloud!

**gcloud-cleanup** takes care of cleaning up resources from the Google Compute Platform.

## Status

Actively running in production.

## How does it fit into the rest of the system

* **Deployment**: Heroku, one instance per google cloud project
* **Google Cloud**: We talk to the Google Compute Platform via its API
* **Job-board** ([github](https://github.com/travis-ci/job-board)):
  We talk to job-board via HTTP to get information about which images
  are still in use.
* **Worker** ([github](https://github.com/travis-ci/worker)): gcloud-cleanup
  shares a redis instance with worker for API call rate limiting

## What does it really do

### Instance cleaning

gcloud-cleanup finds instances that have existed for longer than a certain _cutoff
time_, which can be configured in the `GCLOUD_CLEANUP_INSTANCE_MAX_AGE` environment
variable. These instances are terminated and deleted.

This ensures that instances that failed to terminate are cleaned up.

### Image cleaning

TBD.

### Rate limiting

GCE is not happy if we send them a gazillion API requests. In order to prevent
us from being rate limited, we throttle the amount of requests we make.

## Config parameters

* GCLOUD_CLEANUP_INSTANCE_MAX_AGE:     3h
* GCLOUD_CLEANUP_LOOP_SLEEP:           1s
* GCLOUD_CLEANUP_RATE_LIMIT_DURATION:  1s
* GCLOUD_CLEANUP_RATE_LIMIT_MAX_CALLS: 2
* GCLOUD_CLEANUP_RATE_LIMIT_PREFIX:    rate-limit-gcloud-cleanup
* GCLOUD_CLEANUP_RATE_LIMIT_REDIS_URL: ...
* JOB_BOARD_URL:                       ...
