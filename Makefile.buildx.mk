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

# ==============================================================================
# Build Options

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL}

ifeq (${TAG}, )
  export TAG=ci
endif

# Initialize the NDM DaemonSet variables
# Specify the NDM DaemonSet binary name
NODE_DISK_MANAGER=ndm
# Specify the sub path under ./cmd/ for NDM DaemonSet
BUILD_PATH_NDM=ndm_daemonset
# Name of the image for NDM DaemoneSet
DOCKER_IMAGE_NDM:=${IMAGE_ORG}/node-disk-manager:${TAG}

# Initialize the NDM Operator variables
# Specify the NDM Operator binary name
NODE_DISK_OPERATOR=ndo
# Specify the sub path under ./cmd/ for NDM Operator
BUILD_PATH_NDO=manager
# Name of the image for ndm operator
DOCKER_IMAGE_NDO:=${IMAGE_ORG}/node-disk-operator:${TAG}

# Initialize the NDM Exporter variables
# Specfiy the NDM Exporter binary name
NODE_DISK_EXPORTER=exporter
# Specify the sub path under ./cmd/ for NDM Exporter
BUILD_PATH_EXPORTER=ndm-exporter
# Name of the image for ndm exporter
DOCKER_IMAGE_EXPORTER:=${IMAGE_ORG}/node-disk-exporter:${TAG}

.PHONY: buildx.ndm
buildx.ndm: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-manager binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} BUILDPATH=${BUILD_PATH_NDM} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

# Build ndm docker image with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/
.PHONY: docker.buildx.ndm
docker.buildx.ndm:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" --name container-builder --use;\
	fi
	@docker buildx build --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" \
		-t "$(DOCKER_IMAGE_NDM)" ${DBUILD_ARGS} -f ndm-daemonset.Dockerfile \
		. --push
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDM)"
	@echo

.PHONY: buildx.ndo
buildx.ndo: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-operator binary...'
	@pwd
	@CTLNAME=${NODE_DISK_OPERATOR} BUILDPATH=${BUILD_PATH_NDO} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.ndo
docker.buildx.ndo:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" --name container-builder --use;\
	fi
	@docker buildx build --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" \
		-t "$(DOCKER_IMAGE_NDO)" ${DBUILD_ARGS} -f ndm-operator.Dockerfile \
		. --push
	@echo "--> Build docker image: $(DOCKER_IMAGE_NDO)"
	@echo

.PHONY: buildx.exporter
buildx.exporter: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-exporter binary...'
	@pwd
	@CTLNAME=${NODE_DISK_EXPORTER} BUILDPATH=${BUILD_PATH_EXPORTER} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.exporter
docker.buildx.exporter:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" --name container-builder --use;\
	fi
	@docker buildx build --platform "linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le" \
		-t "$(DOCKER_IMAGE_EXPORTER)" ${DBUILD_ARGS} -f ndm-exporter.Dockerfile \
		. --push
	@echo "--> Build docker image: $(DOCKER_IMAGE_EXPORTER)"
	@echo

.PHONY: install-dep-nonsudo
install-dep-nonsudo:
	@echo "--> Installing external dependencies for building node-disk-manager"
	$(PWD)/build/install-dep.sh

.PHONY: buildx.push.ndm
buildx.push.ndm:
	BUILDX=true DIMAGE=${IMAGE_ORG}/node-disk-manager ./build/push

.PHONY: buildx.push.exporter
buildx.push.exporter:
	BUILDX=true DIMAGE=${IMAGE_ORG}/node-disk-exporter ./build/push

.PHONY: buildx.push.ndo
buildx.push.ndo:
	BUILDX=true DIMAGE=${IMAGE_ORG}/node-disk-operator ./build/push
