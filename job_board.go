package gcloudcleanup

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

type jobBoardImagesResponse struct {
	Data []*jobBoardImageRef `json:"data"`
}

type jobBoardImageRef struct {
	Name string `json:"name"`
}

func makeJobBoardImagesRequest(urlString string) (*jobBoardImagesResponse, error) {
	var responseBody []byte

	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 10 * time.Second
	b.MaxElapsedTime = time.Minute

	err := backoff.Retry(func() (err error) {
		resp, err := http.Get(urlString)

		if err != nil {
			return err
		}
		defer resp.Body.Close()
		responseBody, err = ioutil.ReadAll(resp.Body)
		return
	}, b)

	if err != nil {
		return nil, err
	}

	imageResp := &jobBoardImagesResponse{
		Data: []*jobBoardImageRef{},
	}

	err = json.Unmarshal(responseBody, imageResp)
	if err != nil {
		return nil, err
	}

	return imageResp, nil
}
