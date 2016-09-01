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
  are known.
* **Worker** ([github](https://github.com/travis-ci/worker)): gcloud-cleanup
  shares a redis instance with worker for API call rate limiting

## What does it really do

### Instance cleaning

gcloud-cleanup finds instances matching _name filters_ that have existed for
longer than a certain _cutoff time_ and deletes them.

This ensures that instances that failed to terminate are cleaned up.

Relevant configuration:

- `GCLOUD_CLEANUP_INSTANCE_FILTERS` correspond to _name filters_,
  default `name eq ^testing-gce.*`.
- `GCLOUD_CLEANUP_INSTANCE_MAX_AGE` corresponds to _cutoff time_, default `3h`.

### Image cleaning

gcloud-cleanup queries **Job-board** for all known images matching _name
filters_ with `infra=gce`, then queries **Google Cloud** for all known images
matching _name filters_, and deletes any images in the **Google Cloud** set that
are not also in the **Job-board** set.

This ensures that images unknown to **Job-board** are cleaned up.

Relevant configuration:

- `GCLOUD_CLEANUP_IMAGE_FILTERS` corresponds to _name filters_,
  default `name eq ^travis-ci.*`.

### Rate limiting

GCE is not happy if we send them a gazillion API requests. In order to prevent
us from being rate limited, we throttle the amount of requests we make.

**NOTE**: the `./ratelimit` subpacakage is a vendored copy from
[travis-ci/worker](https://github.com/travis-ci/worker).

## Config parameters

Viewable in [USAGE.txt](./USAGE.txt).
