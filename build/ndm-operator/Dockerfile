# Copyright 2019-2020 The OpenEBS Authors. All rights reserved.
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
#
# This Dockerfile builds ndm operator
#
FROM ghcr.io/openebs/ndm-build-base:20211125 as build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG RELEASE_TAG

ENV GO111MODULE=on \
  CGO_ENABLED=1 \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT} \
  DEBIAN_FRONTEND=noninteractive \
  PATH="/root/go/bin:${PATH}" \
  RELEASE_TAG=${RELEASE_TAG}

WORKDIR /go/src/github.com/openebs/node-disk-manager/

RUN apt-get update && apt-get install -y make git

COPY . .

RUN make NODEPINSTALL=true buildx.ndo

FROM ubuntu

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL

LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="node-disk-operator"
LABEL org.label-schema.description="OpenEBS NDM Operator"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

COPY --from=build /go/src/github.com/openebs/node-disk-manager/bin/ndo /usr/local/bin/ndo

ENTRYPOINT ["/usr/local/bin/ndo"]
