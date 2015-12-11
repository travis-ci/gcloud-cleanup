package gcloudcleanup

/*
__delete_old_images() {
  local regimg_tmp=$(mktemp ${TMPDIR:-/var/tmp}/gcloud-cleanup-regimg.XXXXX)
  local gceimg_tmp=$(mktemp ${TMPDIR:-/var/tmp}/gcloud-cleanup-gceimg.XXXXX)
  local delimg_tmp=$(mktemp ${TMPDIR:-/var/tmp}/gcloud-cleanup-delimg.XXXXX)

  log 'fetching GCE CI images'

  ${GCLOUD_READ_EXE} compute images list \
    --regexp 'travis-ci-.*' \
    --format json \
    --limit ${GCLOUD_CLEANUP_IMAGE_LIMIT} | \
    jq -r '.[] | .name' | sort > ${gceimg_tmp}

  log 'fetching registered CI images'

  local params="infra=gce&name=^travis-ci-.*&fields\[images\]=name"
  params="${params}&limit=${GCLOUD_CLEANUP_IMAGE_LIMIT}"

  set +o errexit
  (
    curl -sSL "${JOB_BOARD_URL}/images?${params}" 2>/dev/null || echo '{"data":[]}'
  ) | \
    jq -r '.data | .[] | .name' | sort | grep -v '^$' > ${regimg_tmp}
  set -o errexit

  if [[ ! -s ${regimg_tmp} ]] ; then
    log 'no registered images?'
    rm -f ${regimg_tmp} ${gceimg_tmp}
    return
  fi

  if [[ ! -s ${gceimg_tmp} ]] ; then
    log 'no GCE images?'
    rm -f ${regimg_tmp} ${gceimg_tmp}
    return
  fi

  set +o errexit
  __set_diff "${regimg_tmp}" "${gceimg_tmp}" 2>/dev/null | \
    grep -v '^$' 2>/dev/null > "${delimg_tmp}"
  set -o errexit

  if [[ -s ${delimg_tmp} ]] ; then
    set +o errexit
    cat "${delimg_tmp}" | \
      xargs -- ${GCLOUD_WRITE_EXE} compute images delete \
        --verbosity=${GCLOUD_VERBOSITY} \
        -q \
        --${GCLOUD_LOG_HTTP}
    set -o errexit
    log 'deleted images' n=$(wc -l ${delimg_tmp} | awk '{ print $1 }')
  fi

  rm -f ${regimg_tmp} ${gceimg_tmp} ${delimg_tmp}
}
*/

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"google.golang.org/api/compute/v1"
)

type imageCleaner struct {
	cs  *compute.Service
	log *logrus.Entry

	projectID   string
	rateLimiter *time.Ticker
	filters     []string
}

func newImageCleaner(cs *compute.Service, log *logrus.Logger,
	rlTick time.Duration, projectID string, filters []string) *imageCleaner {

	return &imageCleaner{
		cs:  cs,
		log: log.WithField("component", "image_cleaner"),

		projectID:   projectID,
		rateLimiter: time.NewTicker(rlTick),
		filters:     filters,
	}
}

func (ic *imageCleaner) Run() error {
	ic.log.WithFields(logrus.Fields{
		"project": ic.projectID,
		"filters": strings.Join(ic.filters, ","),
	}).Info("running image cleanup")

	imageChan := make(chan *compute.Image)
	errChan := make(chan error)

	go ic.fetchImagesToDelete(imageChan, errChan)
	go func() {
		for err := range errChan {
			if err == nil {
				continue
			}
			ic.log.WithField("err", err).Warn("error during image fetch")
		}
	}()

	for image := range imageChan {
		if image == nil {
			return nil
		}

		err := ic.deleteImage(image)

		if err != nil {
			ic.log.WithFields(logrus.Fields{
				"err":   err,
				"image": image.Name,
			}).Warn("failed to delete image")
		}

		ic.log.WithField("image", image.Name).Info("deleted")
	}

	return nil
}

func (ic *imageCleaner) fetchImagesToDelete(imgChan chan *compute.Image, errChan chan error) {
	imgChan <- nil
	errChan <- nil
	return
}

func (ic *imageCleaner) deleteImage(image *compute.Image) error {
	return nil
}

func (ic *imageCleaner) apiRateLimit() {
	ic.log.Debug("waiting for rate limiter tick")
	<-ic.rateLimiter.C
}
