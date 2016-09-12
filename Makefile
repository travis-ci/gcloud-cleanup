include .shellbits.mk
PACKAGE_CHECKOUT := $(shell echo ${PWD})
PACKAGE := github.com/travis-ci/gcloud-cleanup
ALL_PACKAGES := $(PACKAGE) $(PACKAGE)/cmd/...

VERSION_VAR := $(PACKAGE).VersionString
VERSION_VALUE ?= $(shell $(GIT) describe --always --dirty --tags 2>/dev/null)
REV_VAR := $(PACKAGE).RevisionString
REV_VALUE ?= $(shell $(GIT) rev-parse HEAD 2>/dev/null || echo "???")
REV_URL_VAR := $(PACKAGE).RevisionURLString
REV_URL_VALUE ?= https://github.com/travis-ci/gcloud-cleanup/tree/$(shell $(GIT) rev-parse HEAD 2>/dev/null || echo "'???'")
GENERATED_VAR := $(PACKAGE).GeneratedString
GENERATED_VALUE ?= $(shell $(DATE) -u +'%Y-%m-%dT%H:%M:%S%z')
COPYRIGHT_VAR := $(PACKAGE).CopyrightString
COPYRIGHT_VALUE ?= $(shell $(GREP) -i ^copyright LICENSE | $(SED) 's/^[Cc]opyright //')

OS := $(shell $(UNAME) | $(TR) '[:upper:]' '[:lower:]')
ARCH := $(shell $(UNAME) -m | if $(GREP) -q x86_64 ; then echo amd64 ; else $(UNAME) -m ; fi)
GOPATH := $(shell echo $${GOPATH%%:*})
GOBUILD_LDFLAGS ?= \
	-X '$(VERSION_VAR)=$(VERSION_VALUE)' \
	-X '$(REV_VAR)=$(REV_VALUE)' \
	-X '$(REV_URL_VAR)=$(REV_URL_VALUE)' \
	-X '$(GENERATED_VAR)=$(GENERATED_VALUE)' \
	-X '$(COPYRIGHT_VAR)=$(COPYRIGHT_VALUE)'

export GO15VENDOREXPERIMENT

.PHONY: all
all: clean build test coverage.html crossbuild

.PHONY: clean
clean:
	$(RM) $(GOPATH)/bin/gcloud-cleanup
	$(RM) -rv ./build coverage.html coverage.txt
	$(FIND) $(GOPATH)/pkg -wholename "*$(PACKAGE)*.a" -delete

.PHONY: test
test: .test coverage.txt coverage.html

coverage.txt: .test

.PHONY: .test
.test:
	$(GO) test -x -v -cover \
		-coverpkg $(PACKAGE) \
		-coverprofile coverage.txt \
		$(PACKAGE)

coverage.html: coverage.txt
	$(GO) tool cover -html=$^ -o $@

.PHONY: build
build: deps
	$(GO) install -x -ldflags "$(GOBUILD_LDFLAGS)" $(ALL_PACKAGES)

.PHONY: crossbuild
crossbuild: deps
	GOARCH=amd64 GOOS=darwin $(GO) build -o build/darwin/amd64/gcloud-cleanup \
		-ldflags "$(GOBUILD_LDFLAGS)" $(PACKAGE)/cmd/gcloud-cleanup
	GOARCH=amd64 GOOS=linux $(GO) build -o build/linux/amd64/gcloud-cleanup \
		-ldflags "$(GOBUILD_LDFLAGS)" $(PACKAGE)/cmd/gcloud-cleanup

.PHONY: distclean
distclean: clean
	$(RM) vendor/.deps-fetched

.PHONY: deps
deps: vendor/.deps-fetched

.PHONY: prereqs
prereqs:
	$(GO) get github.com/FiloSottile/gvt

.PHONY: copyright
copyright:
	$(SED) -i "s/^Copyright.*Travis CI/Copyright © $(shell date +%Y) Travis CI/" LICENSE

USAGE.txt: build/$(OS)/$(ARCH)/gcloud-cleanup
	source .example.env && \
	$^ --help | sed '/VERSION/ { N; N; d; };s/  *$$//g' >$@

vendor/.deps-fetched:
	$(GVT) rebuild
	$(TOUCH) $@
