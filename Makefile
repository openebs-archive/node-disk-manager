# Specify the name for the binaries
NODE_DISK_MANAGER=ndm

# Build the node-disk-manager image.

build: clean vet fmt ndm version docker

PACKAGES = $(shell go list ./... | grep -v '/vendor/' | grep -v 'integration_test')

NODE_DISK_MANAGER?=ndm

# Determine the arch/os
XC_OS?= $(shell go env GOOS)
XC_ARCH?= $(shell go env GOARCH)
ARCH:=${XC_OS}_${XC_ARCH}

# VERSION is the version of the binary.
VERSION:=$(shell git describe --tags --always)

# TAG is the tag of the docker image
TAG?=$(VERSION)

# IMAGE is the image name of the node-disk-manager docker image.
IMAGE:=openebs/node-disk-manager-${XC_ARCH}:${TAG}

# The ubuntu:16.04 image is being used as base image.
BASEIMAGE:=ubuntu:16.04

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v1

vet:
	go list ./... | grep -v "./vendor/*" | xargs go vet

fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

# Run the bootstrap target once before trying gometalinter in Development environment
golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --deadline=600s ./...

version:
	@echo $(VERSION)

test: 	vet fmt
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done

Dockerfile: Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

header:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@echo

integration-test:
	go test -v github.com/openebs/node-disk-manager/integration_test

ndm:
	@echo '--> Building binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} sh -c "'$(PWD)/hack/build.sh'"
	@echo '--> Built binary.'
	@echo

deps: header
	@echo '--> Resolving dependencies...'
	dep ensure
	@echo '--> Depedencies resolved.'
	@echo

docker: Dockerfile
	@echo "--> Building docker image..."
	@docker build -t "$(IMAGE)" --build-arg ARCH=${ARCH} .
	@echo "--> Build docker image: $(IMAGE)"
	@echo


clean: header
	@echo '--> Cleaning directory...'
	rm -rf bin
	rm -rf ${GOPATH}/bin/${NODE_DISK_MANAGER}
	rm -rf ${GOPATH}/pkg/*
	@echo '--> Done cleaning.'
	@echo

.PHONY: build
