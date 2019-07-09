# Specify the name for the binaries
NODE_DISK_MANAGER=ndm
NODE_DISK_OPERATOR=ndo

# env for specifying that we want to build ndm daemonset
BUILD_PATH_NDM=ndm_daemonset

# env for specifying that we want to build node-disk-operator
BUILD_PATH_NDO=manager

# Build the node-disk-manager image.
build: clean vet fmt shellcheck license-check-go version ndm docker_ndm ndo docker_ndo

NODE_DISK_MANAGER?=ndm
NODE_DISK_OPERATOR?=ndo
BUILD_PATH_NDM?=ndm_daemonset
BUILD_PATH_NDO?=manager

# Determine the arch/os
XC_OS?= $(shell go env GOOS)
XC_ARCH?= $(shell go env GOARCH)
ARCH:=${XC_OS}_${XC_ARCH}

# VERSION is the version of the binary.
VERSION:=$(shell git describe --tags --always)

# IMAGE is the image name of the node-disk-manager docker image.
DOCKER_IMAGE_NDM:=openebs/node-disk-manager-${XC_ARCH}:ci
DOCKER_IMAGE_NDO:=openebs/node-disk-operator-${XC_ARCH}:ci

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

Dockerfile.ndo: ./build/ndm-operator/Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

header:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@echo

integration-test:
	go test -v -timeout 15m github.com/openebs/node-disk-manager/integration_tests/sanity

ndm:
	@echo '--> Building node-disk-manager binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} BUILDPATH=${BUILD_PATH_NDM} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

docker_ndm: Dockerfile.ndm 
	@echo "--> Building docker image for ndm-daemonset..."
	@sudo docker build -t "$(DOCKER_IMAGE_NDM)" --build-arg ARCH=${ARCH} -f Dockerfile.ndm .
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDM)"
	@echo
ndo:
	@echo '--> Building node-disk-operator binary...'
	@pwd
	@CTLNAME=${NODE_DISK_OPERATOR} BUILDPATH=${BUILD_PATH_NDO} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

docker_ndo: Dockerfile.ndo 
	@echo "--> Building docker image for ndm-operator..."
	@sudo docker build -t "$(DOCKER_IMAGE_NDO)" --build-arg ARCH=${ARCH} -f Dockerfile.ndo .
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDO)"
	@echo

deps: header
	@echo '--> Resolving dependencies...'
	dep ensure
	@echo '--> Depedencies resolved.'
	@echo

clean: header
	@echo '--> Cleaning directory...'
	rm -rf bin
	rm -rf ${GOPATH}/bin/${NODE_DISK_MANAGER}
	rm -rf ${GOPATH}/bin/${NODE_DISK_OPERATOR}
	rm -rf ${GOPATH}/pkg/*
	@echo '--> Done cleaning.'
	@echo

license-check-go:
	@echo "--> Checking license header..."
	@licRes=$$(for file in $$(find . -type f -iname '*.go' ! -path './vendor/*' ) ; do \
               awk 'NR<=3' $$file | grep -Eq "(Copyright|generated|GENERATED)" || echo $$file; \
       done); \
       if [ -n "$${licRes}" ]; then \
               echo "license header checking failed:"; echo "$${licRes}"; \
               exit 1; \
       fi
	@echo "--> Done checking license."
	@echo

.PHONY: build
