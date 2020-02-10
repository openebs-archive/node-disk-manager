#!/bin/bash

# This script is used to install external dependencies that are
# used for building NDM as well as checking

set -e
# udev is required for building NDM.
sudo apt-get install --yes libudev-dev
pushd .
cd ..
# we need openSeaChest repo to build node-disk-manager
git clone --recursive --branch Release-19.06.02 https://github.com/openebs/openSeaChest.git
cd openSeaChest/Make/gcc
make release
cd ../../
sudo cp opensea-common/Make/gcc/lib/libopensea-common.a /usr/lib
sudo cp opensea-operations/Make/gcc/lib/libopensea-operations.a /usr/lib
sudo cp opensea-transport/Make/gcc/lib/libopensea-transport.a /usr/lib
popd