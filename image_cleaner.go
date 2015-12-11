package gcloudcleanup

import (
	"fmt"
	"net/url"
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
	imageLimit  int

	noop bool
}

func newImageCleaner(cs *compute.Service, log *logrus.Logger,
	rlTick time.Duration, projectID, jobBoardURL string,
	imageLimit int, filters []string, noop bool) *imageCleaner {

	return &imageCleaner{
		cs:  cs,
		log: log.WithField("component", "image_cleaner"),

		projectID:   projectID,
		jobBoardURL: jobBoardURL,
		rateLimiter: time.NewTicker(rlTick),
		imageLimit:  imageLimit,
		filters:     filters,

		noop: noop,
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

	ic.log.WithField("count", len(registeredImages)).Debug("fetched registered images")

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
	images := map[string]bool{}
	nameFilter := ""

	for _, filter := range ic.filters {
		if !strings.HasPrefix(filter, "name eq") {
			continue
		}

		nameFilter = strings.Replace(filter, "name eq", "", -1)
		nameFilter = strings.Trim(strings.TrimSpace(nameFilter), "'\"")
	}

	if nameFilter == "" {
		nameFilter = "^travis-ci.*"
	}

	qs := url.Values{}
	qs.Set("infra", "gce")
	qs.Set("fields[images]", "name")
	qs.Set("name", nameFilter)
	qs.Set("limit", fmt.Sprintf("%v", ic.imageLimit))

	u, err := url.Parse(ic.jobBoardURL)
	u.Path = "/images"
	u.RawQuery = qs.Encode()

	if err != nil {
		return images, err
	}

	imageResp, err := makeJobBoardImagesRequest(u.String())
	if err != nil {
		return images, err
	}

	if len(imageResp.Data) == 0 {
		return images, err
	}

	for _, imgRef := range imageResp.Data {
		images[imgRef.Name] = true
	}

	return images, nil
}

func (ic *imageCleaner) fetchImagesToDelete(registeredImages map[string]bool,
	imgChan chan *compute.Image, errChan chan error) {

	listCall := ic.cs.Images.List(ic.projectID)
	for _, filter := range ic.filters {
		listCall.Filter(filter)
	}

	pageTok := ""

	for {
		if pageTok != "" {
			listCall.PageToken(pageTok)
		}

		ic.apiRateLimit()
		ic.log.WithField("page_token", pageTok).Debug("fetching images list")
		resp, err := listCall.Do()

		if err != nil {
			errChan <- err
			continue
		}

		for _, image := range resp.Items {
			if _, ok := registeredImages[image.Name]; !ok {
				ic.log.WithField("image", image.Name).Debug("sending image for deletion")

				imgChan <- image
				continue
			}

			ic.log.WithField("image", image.Name).Debug("skipping image")
		}

		if resp.NextPageToken == "" {
			ic.log.Debug("no next page, breaking out of loop")
			imgChan <- nil
			errChan <- nil
			return
		}

		ic.log.Debug("continuing to next page")
		pageTok = resp.NextPageToken
	}
}

func (ic *imageCleaner) deleteImage(image *compute.Image) error {
	if ic.noop {
		ic.log.WithField("image", image.Name).Debug("not really deleting image")
		return nil
	}

	ic.apiRateLimit()
	_, err := ic.cs.Images.Delete(ic.projectID, image.Name).Do()
	return err
}

func (ic *imageCleaner) apiRateLimit() {
	ic.log.Debug("waiting for rate limiter tick")
	<-ic.rateLimiter.C
}
