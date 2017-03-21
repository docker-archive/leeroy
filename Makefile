.PHONY: clean all fmt vet lint build test install static
PREFIX?=$(shell pwd)
BUILDTAGS=

PROJECT := github.com/docker/leeroy
VENDOR := vendor

# Variable to get the current version.
VERSION := $(shell cat VERSION)

# Variable to set the current git commit.
GITCOMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
GITCOMMIT := $(GITCOMMIT)-dirty
endif

LDFLAGS := ${LDFLAGS} \
	-X $(PROJECT)/version.GitCommit=${GITCOMMIT} \
	-X $(PROJECT)/version.Version=${VERSION} \

all: clean build fmt lint test vet install

build:
	@echo "+ $@"
	go build -tags "$(BUILDTAGS)" -ldflags "${LDFLAGS}" .

static:
	@echo "+ $@"
	CGO_ENABLED=0 go build -tags "$(BUILDTAGS) static_build" \
		-ldflags "-w -extldflags -static ${LDFLAGS}" -o leeroy .

fmt:
	@echo "+ $@"
	@gofmt -s -l . | grep -v $(VENDOR) | tee /dev/stderr

lint:
	@echo "+ $@"
	@golint ./... | grep -v $(VENDOR) | tee /dev/stderr

test: fmt lint vet
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v $(VENDOR))

vet:
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v $(VENDOR))

clean:
	@echo "+ $@"
	@$(RM) leeroy

install:
	@echo "+ $@"
	@go install .
