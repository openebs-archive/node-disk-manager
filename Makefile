# Copyright 2018-2020 The OpenEBS Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Default behaviour is not to use BUILDX until the Travis workflow is deprecated.
BUILDX:=false

# ==============================================================================
# Build Options

# set the shell to bash in case some environments use sh
SHELL:=/bin/bash

# VERSION is the version of the binary.
VERSION:=$(shell git describe --tags --always)

# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS

ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH

ARCH:=${XC_OS}_${XC_ARCH}
export ARCH

ifeq (${BASE_DOCKER_IMAGEARM64}, )
  BASE_DOCKER_IMAGEARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGEARM64
endif

ifeq (${BASEIMAGE}, )
ifeq ($(ARCH),linux_arm64)
  BASEIMAGE:=${BASE_DOCKER_IMAGEARM64}
else
  # The ubuntu:16.04 image is being used as base image.
  BASEIMAGE:=ubuntu:16.04
endif
endif
export BASEIMAGE

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in
# the `build/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique
# repository under an organization.
# Putting all this together, an unique uri for a given
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization
# under which images should be pushed.
# By default the organization name is `openebs`.

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG="openebs"
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/node-disk-manager"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg ARCH=${ARCH}


# Initialize the NDM DaemonSet variables
# Specify the NDM DaemonSet binary name
NODE_DISK_MANAGER=ndm
# Specify the sub path under ./cmd/ for NDM DaemonSet
BUILD_PATH_NDM=ndm_daemonset
# Name of the image for NDM DaemoneSet
DOCKER_IMAGE_NDM:=${IMAGE_ORG}/node-disk-manager-${XC_ARCH}:ci

# Initialize the NDM Operator variables
# Specify the NDM Operator binary name
NODE_DISK_OPERATOR=ndo
# Specify the sub path under ./cmd/ for NDM Operator
BUILD_PATH_NDO=manager
# Name of the image for ndm operator
DOCKER_IMAGE_NDO:=${IMAGE_ORG}/node-disk-operator-${XC_ARCH}:ci

# Initialize the NDM Exporter variables
# Specfiy the NDM Exporter binary name
NODE_DISK_EXPORTER=exporter
# Specify the sub path under ./cmd/ for NDM Exporter
BUILD_PATH_EXPORTER=ndm-exporter
# Name of the image for ndm exporter
DOCKER_IMAGE_EXPORTER:=${IMAGE_ORG}/node-disk-exporter-${XC_ARCH}:ci

# Compile binaries and build docker images
.PHONY: build
build: clean build.common docker.ndm docker.ndo docker.exporter

.PHONY: build.common
build.common: license-check-go version

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/mitchellh/gox

# Bootstrap the build by downloading additional tools
.PHONY: bootstrap
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done

.PHONY: install-dep
install-dep:
	@echo "--> Installing external dependencies for building node-disk-manager"
	@sudo $(PWD)/build/install-dep.sh

.PHONY: install-test-infra
install-test-infra:
	@echo "--> Installing test infra for running integration tests"
	# installing test infrastructure is dependent on the platform
	$(PWD)/build/install-test-infra.sh ${XC_ARCH}

.PHONY: header
header:
	@echo "----------------------------"
	@echo "--> node-disk-manager       "
	@echo "----------------------------"
	@echo

# -composite: avoid "literal copies lock value from fakePtr"
.PHONY: vet
vet:
	go list ./... | grep -v "./vendor/*" | xargs go vet -composites

.PHONY: fmt
fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

# shellcheck target for checking shell scripts linting
.PHONY: shellcheck
shellcheck: getshellcheck
	find . -type f -name "*.sh" | grep -v "./vendor/*" | xargs /tmp/shellcheck-stable/shellcheck

.PHONY: getshellcheck
getshellcheck:
	wget -c 'https://github.com/koalaman/shellcheck/releases/download/stable/shellcheck-stable.linux.x86_64.tar.xz' --no-check-certificate -O - | tar -xvJ -C /tmp/
.PHONY: version
version:
	@echo $(VERSION)

.PHONY: test
test: 	vet fmt
	@echo "--> Running go test";
	$(PWD)/build/test.sh ${XC_ARCH}

.PHONY: integration-test
integration-test:
	@echo "--> Running integration test"
	$(PWD)/build/integration-test.sh ${XC_ARCH}

.PHONY: Dockerfile.ndm
Dockerfile.ndm: ./build/ndm-daemonset/Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

.PHONY: Dockerfile.ndo
Dockerfile.ndo: ./build/ndm-operator/Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

.PHONY: Dockerfile.exporter
Dockerfile.exporter: ./build/ndm-exporter/Dockerfile.in
	sed -e 's|@BASEIMAGE@|$(BASEIMAGE)|g' $< >$@

.PHONY: build.ndm
build.ndm:
	@echo '--> Building node-disk-manager binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} BUILDPATH=${BUILD_PATH_NDM} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.ndm
docker.ndm: build.ndm Dockerfile.ndm
	@echo "--> Building docker image for ndm-daemonset..."
	@sudo docker build -t "$(DOCKER_IMAGE_NDM)" ${DBUILD_ARGS} -f Dockerfile.ndm .
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDM)"
	@echo

.PHONY: build.ndo
build.ndo:
	@echo '--> Building node-disk-operator binary...'
	@pwd
	@CTLNAME=${NODE_DISK_OPERATOR} BUILDPATH=${BUILD_PATH_NDO} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.ndo
docker.ndo: build.ndo Dockerfile.ndo
	@echo "--> Building docker image for ndm-operator..."
	@sudo docker build -t "$(DOCKER_IMAGE_NDO)" ${DBUILD_ARGS} -f Dockerfile.ndo .
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDO)"
	@echo

.PHONY: build.exporter
build.exporter:
	@echo '--> Building node-disk-exporter binary...'
	@pwd
	@CTLNAME=${NODE_DISK_EXPORTER} BUILDPATH=${BUILD_PATH_EXPORTER} sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.exporter
docker.exporter: build.exporter Dockerfile.exporter
	@echo "--> Building docker image for ndm-exporter..."
	@sudo docker build -t "$(DOCKER_IMAGE_EXPORTER)" ${DBUILD_ARGS} -f Dockerfile.exporter .
	@echo "--> Build docker image: $(DOCKER_IMAGE_EXPORTER)"
	@echo

.PHONY: deps
deps: header
	@echo '--> Resolving dependencies...'
	go mod tidy
	go mod verify
	go mod vendor
	@echo '--> Depedencies resolved.'
	@echo

.PHONY: clean
clean: header
	@echo '--> Cleaning directory...'
	rm -rf bin
	rm -rf ${GOPATH}/bin/${NODE_DISK_MANAGER}
	rm -rf ${GOPATH}/bin/${NODE_DISK_OPERATOR}
	rm -rf ${GOPATH}/bin/${NODE_DISK_EXPORTER}
	rm -rf Dockerfile.ndm
	rm -rf Dockerfile.ndo
	rm -rf Dockerfile.exporter
	@echo '--> Done cleaning.'
	@echo

.PHONY: license-check-go
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

.PHONY: push
push:
	DIMAGE=${IMAGE_ORG}/node-disk-manager-${XC_ARCH} ./build/push;
	DIMAGE=${IMAGE_ORG}/node-disk-operator-${XC_ARCH} ./build/push;
	DIMAGE=${IMAGE_ORG}/node-disk-exporter-${XC_ARCH} ./build/push;

#-----------------------------------------------------------------------------
# Target: docker.buildx.ndm docker.buildx.ndo docker.buildx.exporter
#-----------------------------------------------------------------------------
include Makefile.buildx.mk
