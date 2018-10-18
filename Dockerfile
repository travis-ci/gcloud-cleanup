FROM golang:1.11 as builder
MAINTAINER Travis CI GmbH <support+gcloud-cleanup-docker-image@travis-ci.org>

RUN go get -u github.com/FiloSottile/gvt

COPY . /go/src/github.com/travis-ci/gcloud-cleanup
WORKDIR /go/src/github.com/travis-ci/gcloud-cleanup
RUN make deps
ENV CGO_ENABLED 0
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl bash

COPY --from=builder /go/bin/gcloud-cleanup /usr/local/bin/gcloud-cleanup

ENTRYPOINT ["/usr/local/bin/gcloud-cleanup"]
