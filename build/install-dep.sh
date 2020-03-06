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
if [ "$ARCH" == "arm" ]; then
    # Until #34 is merged, a new release is cut, and it is usable by openebs NDM
    sed -ne "s@\\$(MAKE) -C@\$(MAKE) \$(MAKEFLAG) -C@g" -i openSeaChest/Make/gcc

    # Enable cross compilation (needed until Travis CI gets arm32 support)
    apt-get -y -qq install crossbuild-essential-armhf 
    export CC=arm-linux-gnueabihf-gcc AR=arm-linux-gnueabihf-ar LD=arm-linux-gnueabihf-ld STRIP=arm-linux-gnueabihf-strip
fi
cd openSeaChest/Make/gcc
make release
cd ../../
sudo cp opensea-common/Make/gcc/lib/libopensea-common.a /usr/lib
sudo cp opensea-operations/Make/gcc/lib/libopensea-operations.a /usr/lib
sudo cp opensea-transport/Make/gcc/lib/libopensea-transport.a /usr/lib
popd
