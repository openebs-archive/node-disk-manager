#!/usr/bin/env bash
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
#
# This script builds the application from source for multiple platforms.
set -e

# Get the git commit
getGitCommit()
{
    if [ -f "$GOPATH"/src/github.com/openebs/node-disk-manager/GITCOMMIT ];
    then
	GIT_COMMIT="$(cat "$GOPATH"/src/github.com/openebs/node-disk-manager/GITCOMMIT)"
    else
	GIT_COMMIT="$(git rev-parse HEAD)"
    fi
    echo "${GIT_COMMIT}"
}

# Delete the old contents
deleteOldContents(){
    rm -rf bin/*
    mkdir -p bin/

    echo "Successfully deleted old bin contents"
}

# Move all the compiled things to the $GOPATH/bin
moveCompiled(){
    GOPATH=${GOPATH:-$(go env GOPATH)}
    case $(uname) in
        CYGWIN*)
            GOPATH="$(cygpath "$GOPATH")"
            ;;
    esac
    OLDIFS=$IFS
    IFS=: MAIN_GOPATH=("$GOPATH")
    IFS=$OLDIFS

    # Create the gopath bin if not already available
    mkdir -p "${MAIN_GOPATH[*]}"/bin/

    # Copy our OS/Arch to ${MAIN_GOPATH}/bin/ directory
    DEV_PLATFORM="./bin/$(go env GOOS)_$(go env GOARCH)"
    DEV_PLATFORM_OUTPUT=$(find "${DEV_PLATFORM}" -mindepth 1 -maxdepth 1 -type f)
    for F in ${DEV_PLATFORM_OUTPUT}; do
        cp "${F}" bin/
        cp "${F}" "${MAIN_GOPATH[*]}"/bin/
    done

    echo "Moved all the compiled things successfully to :${MAIN_GOPATH[*]}/bin/"
}

moveCompiledBuildx(){
    GOPATH=${GOPATH:-$(go env GOPATH)}
    OLDIFS=$IFS
    IFS=: MAIN_GOPATH=("$GOPATH")
    IFS=$OLDIFS

    # Create the gopath bin if not already available
    mkdir -p "${MAIN_GOPATH[*]}"/bin/

    # Copy our OS/Arch to ${MAIN_GOPATH}/bin/ directory
    DEV_PLATFORM="./bin"
    cp -r "${DEV_PLATFORM}" "${MAIN_GOPATH[*]}"/bin/

    echo "Moved all the compiled things successfully to :${MAIN_GOPATH[*]}/bin/"
}

# Buildx
buildx(){
    output_name="bin/$CTLNAME"
    echo "Building for: ${GOOS} ${GOARCH}"
    go build \
        -ldflags="-X github.com/openebs/node-disk-manager/pkg/version.GitCommit=${GIT_COMMIT} \
        -X main.CtlName='${CTLNAME}' \
        -X github.com/openebs/node-disk-manager/pkg/version.Version=${VERSION}" \
        -o "$output_name" ./cmd/"$BUILDPATH"

    echo "Buildx Successfully built: ${CTLNAME}"
}

# Build
build(){
    for GOOS in "${XC_OSS[@]}"
    do
        for GOARCH in "${XC_ARCHS[@]}"
        do
            UNDERSCORE="_"
            output_name="bin/${GOOS}${UNDERSCORE}${GOARCH}/$CTLNAME"

            if [ "$GOOS" = "windows" ]; then
                output_name+='.exe'
            fi
            echo "Building for: ${GOOS} ${GOARCH}"
            go build \
                -ldflags="-X github.com/openebs/node-disk-manager/pkg/version.GitCommit=${GIT_COMMIT} \
                -X main.CtlName='${CTLNAME}' \
                -X github.com/openebs/node-disk-manager/pkg/version.Version=${VERSION}" \
                -o "$output_name" ./cmd/"$BUILDPATH"
        done
    done
    echo "Successfully built: ${CTLNAME}"
}

# Main script starts here .......
export CGO_ENABLED=1

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[*]}"

while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that directory
cd "$DIR"

# Get the git commit
GIT_COMMIT=$(getGitCommit)

# Get the version details. By default set as ci.
VERSION="ci"

if [ -n "${RELEASE_TAG}" ] ;
then
    # When github is tagged with a release, then github action release workflow
    # will set the release tag in env RELEASE_TAG
    VERSION="${RELEASE_TAG}"
fi;


# Determine the arch/os combos we're building for
#XC_ARCH=${XC_ARCH:-"amd64"}
#XC_OS=${XC_OS:-"linux"}

XC_ARCHS=("${XC_ARCH// / }")
XC_OSS=("${XC_OS// / }")

echo "==> Removing old bin contents..."
deleteOldContents

# If its dev mode, only build for ourself
if [[ -n "${NDM_AGENT_DEV}" ]]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

# Build!
echo "==> Building ${CTLNAME} ..."
if [[ ${BUILDX} ]]; then
    buildx
    # Move all the compiled things to the $GOPATH/bin
    moveCompiledBuildx
else
    build
    # Move all the compiled things to the $GOPATH/bin
    moveCompiled
fi

ls -hl bin
