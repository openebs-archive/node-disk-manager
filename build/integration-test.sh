#!/bin/bash

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
