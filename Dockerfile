FROM golang:1.11 as builder
MAINTAINER Travis CI GmbH <support+gcloud-cleanup-docker-image@travis-ci.org>

COPY . /tmp/gcloud-cleanup
WORKDIR /tmp/gcloud-cleanup
ENV CGO_ENABLED 0
RUN make build crossbuild

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl bash

COPY --from=builder /tmp/gcloud-cleanup/build/linux/amd64/gcloud-cleanup /usr/local/bin/gcloud-cleanup

CMD ["/usr/local/bin/gcloud-cleanup"]
