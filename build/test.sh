#!/usr/bin/env bash
#
# This script runs tests and generates a report file.

set -e

# architecute on which tests need to be run
ARCH=$1

if [ -z "$ARCH" ]; then
  echo "platform not specified for running tests. Exiting."
  exit 1
fi

# currently tests are run only for amd64
if [ "$ARCH" != "amd64" ]; then
  exit 0
fi

echo "" > coverage.txt
PACKAGES=$(go list ./... | grep -v '/vendor/\|/pkg/apis/\|/pkg/client/\|integration_test')
for d in $PACKAGES; do
	go test -coverprofile=profile.out -covermode=atomic "$d"
	if [ -f profile.out ]; then
		cat profile.out >> coverage.txt
		rm profile.out
	fi
done
