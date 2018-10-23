package gcloudcleanup

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/trace"
	googlecloudtrace "cloud.google.com/go/trace"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type gceAccountJSON struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

func buildGoogleComputeService(accountJSON string) (*compute.Service, error) {
	if accountJSON == "" {
		client, err := google.DefaultClient(context.TODO(), compute.DevstorageFullControlScope, compute.ComputeScope)
		if err != nil {
			return nil, errors.Wrap(err, "could not build default client")
		}
		return compute.New(client)
	}

	a, err := loadGoogleAccountJSON(accountJSON)
	if err != nil {
		return nil, err
	}

	config := jwt.Config{
		Email:      a.ClientEmail,
		PrivateKey: []byte(a.PrivateKey),
		Scopes: []string{
			compute.DevstorageFullControlScope,
			compute.ComputeScope,
		},
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: &ochttp.Transport{},
	})

	client := config.Client(ctx)

	cs, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	cs.UserAgent = "gcloud-cleanup"

	return cs, nil
}

func buildGoogleStorageClient(ctx context.Context, accountJSON string) (*storage.Client, error) {
	if accountJSON == "" {
		creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)
		if err != nil {
			return nil, errors.Wrap(err, "could not build default client")
		}
		return storage.NewClient(ctx, option.WithCredentials(creds))
	}

	credBytes, err := loadBytes(accountJSON)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(ctx, credBytes, storage.ScopeReadWrite)
	if err != nil {
		return nil, err
	}

	return storage.NewClient(ctx, option.WithCredentials(creds))
}

func buildGoogleCloudCredentials(ctx context.Context, accountJSON string) (*google.Credentials, error) {
	if accountJSON == "" {
		creds, err := google.FindDefaultCredentials(ctx, googlecloudtrace.ScopeTraceAppend)
		return creds, errors.Wrap(err, "could not build default client")
	}

	credBytes, err := loadBytes(accountJSON)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(ctx, credBytes, trace.ScopeTraceAppend)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func loadGoogleAccountJSON(filenameOrJSON string) (*gceAccountJSON, error) {
	bytes, err := loadBytes(filenameOrJSON)
	if err != nil {
		return nil, err
	}

	a := &gceAccountJSON{}
	err = json.Unmarshal(bytes, a)
	return a, err
}

func loadBytes(filenameOrJSON string) ([]byte, error) {
	var (
		bytes []byte
		err   error
	)

	if strings.HasPrefix(strings.TrimSpace(filenameOrJSON), "{") {
		bytes = []byte(filenameOrJSON)
	} else {
		bytes, err = ioutil.ReadFile(filenameOrJSON)
		if err != nil {
			return nil, err
		}
	}

	return bytes, nil
}
