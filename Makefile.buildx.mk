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

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Name of the multiarch image for NDM DaemoneSet
DOCKERX_IMAGE_NDM:=${IMAGE_ORG}/node-disk-manager:${TAG}

# Name of the multiarch image for ndm operator
DOCKERX_IMAGE_NDO:=${IMAGE_ORG}/node-disk-operator:${TAG}

# Name of the multiarch image for ndm exporter
DOCKERX_IMAGE_EXPORTER:=${IMAGE_ORG}/node-disk-exporter:${TAG}

# Build ndm docker image with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/

.PHONY: docker.buildx
docker.buildx:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	@docker buildx build --platform ${PLATFORMS} \
		-t "$(DOCKERX_IMAGE_NAME)" ${DBUILD_ARGS} -f $(PWD)/build/$(COMPONENT)/Dockerfile \
		. ${PUSH_ARG}
	@echo "--> Build docker image: $(DOCKERX_IMAGE_NAME)"
	@echo

.PHONY: buildx.ndm
buildx.ndm: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-manager binary...'
	@pwd
	@CTLNAME=${NODE_DISK_MANAGER} BUILDPATH=${BUILD_PATH_NDM} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.ndm
docker.buildx.ndm: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_NDM)
docker.buildx.ndm: COMPONENT=ndm-daemonset
docker.buildx.ndm: docker.buildx

.PHONY: buildx.ndo
buildx.ndo: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-operator binary...'
	@pwd
	@CTLNAME=${NODE_DISK_OPERATOR} BUILDPATH=${BUILD_PATH_NDO} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.ndo
docker.buildx.ndo: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_NDO)
docker.buildx.ndo: COMPONENT=ndm-operator
docker.buildx.ndo: docker.buildx

.PHONY: buildx.exporter
buildx.exporter: bootstrap install-dep-nonsudo clean build.common
	@echo '--> Building node-disk-exporter binary...'
	@pwd
	@CTLNAME=${NODE_DISK_EXPORTER} BUILDPATH=${BUILD_PATH_EXPORTER} BUILDX=true sh -c "'$(PWD)/build/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.exporter
docker.buildx.exporter: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_EXPORTER)
docker.buildx.exporter: COMPONENT=ndm-exporter
docker.buildx.exporter: docker.buildx

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
