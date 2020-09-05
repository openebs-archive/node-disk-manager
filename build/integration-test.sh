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

set -e

# architecute on which integration tests need to be run
ARCH=$1

if [ -z "$ARCH" ]; then
  echo "platform not specified for running tests. Exiting."
  exit 1
fi

# currently integration tests are run only for amd64
if [ "$ARCH" != "amd64" ]; then
  exit 0
fi

go test -v -timeout 20m github.com/openebs/node-disk-manager/integration_tests/sanity
