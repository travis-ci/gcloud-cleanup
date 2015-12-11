PACKAGE_CHECKOUT := $(shell echo ${PWD})
PACKAGE := github.com/travis-ci/gcloud-cleanup
ALL_PACKAGES := $(PACKAGE) $(PACKAGE)/cmd/...

VERSION_VAR := $(PACKAGE).VersionString
VERSION_VALUE ?= $(shell git describe --always --dirty --tags 2>/dev/null)
REV_VAR := $(PACKAGE).RevisionString
REV_VALUE ?= $(shell git rev-parse HEAD 2>/dev/null || echo "'???'")
REV_URL_VAR := $(PACKAGE).RevisionURLString
REV_URL_VALUE ?= https://github.com/travis-ci/gcloud-cleanup/tree/$(shell git rev-parse HEAD 2>/dev/null || echo "'???'")
GENERATED_VAR := $(PACKAGE).GeneratedString
GENERATED_VALUE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%S%z')
COPYRIGHT_VAR := $(PACKAGE).CopyrightString
COPYRIGHT_VALUE ?= $(shell grep -i ^copyright LICENSE | sed 's/^[Cc]opyright //')

FIND ?= find
GO ?= go
GOXC ?= goxc
GVT ?= gvt
TOUCH ?= touch
XARGS ?= xargs

GOPATH := $(shell echo $${GOPATH%%:*})
GOBUILD_LDFLAGS ?= \
	-X '$(VERSION_VAR)=$(VERSION_VALUE)' \
	-X '$(REV_VAR)=$(REV_VALUE)' \
	-X '$(REV_URL_VAR)=$(REV_URL_VALUE)' \
	-X '$(GENERATED_VAR)=$(GENERATED_VALUE)' \
	-X '$(COPYRIGHT_VAR)=$(COPYRIGHT_VALUE)'

export GO15VENDOREXPERIMENT

.PHONY: all
all: clean crossbuild

.PHONY: clean
clean:
	$(RM) $(GOPATH)/bin/gcloud-cleanup
	$(RM) -rv ./build
	$(FIND) $(GOPATH)/pkg -wholename "*$(PACKAGE)*a" | $(XARGS) $(RM) -v

.PHONY: crossbuild
crossbuild: .crossdeps deps
	$(GOXC) -pv=$(VERSION_VALUE) -build-ldflags "$(GOBUILD_LDFLAGS)"

.crossdeps:
	GOROOT_BOOTSTRAP=$(GOROOT) $(GOXC) -t
	touch $@

.PHONY: distclean
distclean: clean
	$(RM) vendor/.deps-fetched

.PHONY: deps
deps: vendor/.deps-fetched

.PHONY: prereqs
prereqs:
	$(GO) get github.com/FiloSottile/gvt
	$(GO) get github.com/laher/goxc

vendor/.deps-fetched:
	$(GVT) rebuild
	touch $@
