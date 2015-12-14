DATE ?= date
FIND ?= find
GIT ?= git
GO ?= go
GOXC ?= goxc
GREP ?= grep
GVT ?= gvt
SED ?= sed
TOUCH ?= touch
TR ?= tr
UNAME ?= uname
XARGS ?= xargs

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

.PHONY: heroku-bin
heroku-bin:
	$(GREP) worker Procfile
	./build/$(OS)/$(ARCH)/gcloud-cleanup --version

.PHONY: all
all: clean test crossbuild

.PHONY: clean
clean:
	$(RM) $(GOPATH)/bin/gcloud-cleanup
	$(RM) -rv ./build
	$(FIND) $(GOPATH)/pkg -wholename "*$(PACKAGE)*a" | $(XARGS) $(RM) -v

.PHONY: test
test:
	$(GO) test -x -v -cover \
		-coverpkg $(PACKAGE) \
		-coverprofile package.coverprofile \
		$(PACKAGE)

.PHONY: build
build: deps
	$(GO) install -x -ldflags "$(GOBUILD_LDFLAGS)" $(ALL_PACKAGES)

.PHONY: crossbuild
crossbuild: .crossdeps deps
	$(GOXC) -pv=$(VERSION_VALUE) -build-ldflags "$(GOBUILD_LDFLAGS)"

.crossdeps:
	GOROOT_BOOTSTRAP=$(GOROOT) $(GOXC) -t
	$(TOUCH) $@

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
	$(TOUCH) $@
