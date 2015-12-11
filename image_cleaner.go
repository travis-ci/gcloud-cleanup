package gcloudcleanup

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
	jobBoardURL string
	rateLimiter *time.Ticker
	filters     []string
}

func newImageCleaner(cs *compute.Service, log *logrus.Logger,
	rlTick time.Duration, projectID, jobBoardURL string,
	filters []string) *imageCleaner {

	return &imageCleaner{
		cs:  cs,
		log: log.WithField("component", "image_cleaner"),

		projectID:   projectID,
		jobBoardURL: jobBoardURL,
		rateLimiter: time.NewTicker(rlTick),
		filters:     filters,
	}
}

func (ic *imageCleaner) Run() error {
	ic.log.WithFields(logrus.Fields{
		"project": ic.projectID,
		"filters": strings.Join(ic.filters, ","),
	}).Info("running image cleanup")

	registeredImages, err := ic.fetchRegisteredImages()
	if err != nil {
		return err
	}

	if len(registeredImages) == 0 {
		ic.log.Warn("no registered images?")
		return nil
	}

	imageChan := make(chan *compute.Image)
	errChan := make(chan error)

	go ic.fetchImagesToDelete(registeredImages, imageChan, errChan)
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

func (ic *imageCleaner) fetchRegisteredImages() (map[string]bool, error) {
	return map[string]bool{}, nil
}

func (ic *imageCleaner) fetchImagesToDelete(registeredImages map[string]bool,
	imgChan chan *compute.Image, errChan chan error) {

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
