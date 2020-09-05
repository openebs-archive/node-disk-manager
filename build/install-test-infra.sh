#!/bin/bash
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

# Test infrastructure for running integration tests on NDM.
# Currently minikube is used to run the integration tests. Since
# minikube is available only on amd64, integration tests can be run
# only on that platform

ARCH=$1

if [ -z "$ARCH" ]; then
  echo "Test Infra platform not specified. Exiting. "
  exit 1
fi

if [ "$ARCH" == "amd64" ]; then
  curl -Lo minikube https://storage.googleapis.com/minikube/releases/v1.0.0/minikube-linux-amd64
  sudo chmod +x minikube
  sudo mv minikube /usr/local/bin/
fi
