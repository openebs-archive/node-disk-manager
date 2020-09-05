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

# This script is used to install external dependencies that are
# used for building NDM as well as checking

set -e

# update the packages
apt-get update -y

# udev and blkid is required for building NDM.
apt-get install --yes libudev-dev libblkid-dev
pushd .
cd ..
# we need openSeaChest repo to build node-disk-manager
git clone --recursive --branch Release-19.06.02 https://github.com/openebs/openSeaChest.git
cd openSeaChest/Make/gcc
make release
cd ../../
cp opensea-common/Make/gcc/lib/libopensea-common.a /usr/lib
cp opensea-operations/Make/gcc/lib/libopensea-operations.a /usr/lib
cp opensea-transport/Make/gcc/lib/libopensea-transport.a /usr/lib
popd
