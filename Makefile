# Specify the name for the binaries
NODE_DISK_MANAGER=ndm

# env for specifying that we want to build ndm daemonset
BUILD_PATH_NDM=ndm_daemonset

# Build the node-disk-manager image.

build: clean vet fmt shellcheck ndm version docker_ndm

NODE_DISK_MANAGER?=ndm
BUILD_PATH_NDM?=ndm_daemonset

# Determine the arch/os
XC_OS?= $(shell go env GOOS)
XC_ARCH?= $(shell go env GOARCH)
ARCH:=${XC_OS}_${XC_ARCH}

# VERSION is the version of the binary.
VERSION:=$(shell git describe --tags --always)

# IMAGE is the image name of the node-disk-manager docker image.
IMAGE:=openebs/node-disk-manager-${XC_ARCH}:ci

# The ubuntu:16.04 image is being used as base image.
BASEIMAGE:=ubuntu:16.04

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	gopkg.in/alecthomas/gometalinter.v1

# -composite: avoid "literal copies lock value from fakePtr"
vet:
	go list ./... | grep -v "./vendor/*" | xargs go vet -composites

fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

# Run the bootstrap target once before trying gometalinter in Development environment
golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --deadline=600s ./...

# shellcheck target for checking shell scripts linting
shellcheck: getshellcheck
	find . -type f -name "*.sh" | grep -v "./vendor/*" | xargs /tmp/shellcheck-latest/shellcheck

getshellcheck:
	wget -c 'https://goo.gl/ZzKHFv' -O - | tar -xvJ -C /tmp/

version:
	@echo $(VERSION)

test: 	vet fmt
	@echo "--> Running go test";
	$(PWD)/build/test.sh

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done

Dockerfile.ndm: ./build/ndm-daemonset/Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

header:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@echo

integration-test:
	go test -v github.com/openebs/node-disk-manager/integration_test

ndm:
	@echo '--> Building node-disk-manager binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} BUILDPATH=${BUILD_PATH_NDM} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

deps: header
	@echo '--> Resolving dependencies...'
	dep ensure
	@echo '--> Depedencies resolved.'
	@echo

docker_ndm: Dockerfile.ndm 
	@echo "--> Building docker image for ndm-daemonset..."
	@sudo docker build -t "$(IMAGE)" --build-arg ARCH=${ARCH} -f Dockerfile.ndm .
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
