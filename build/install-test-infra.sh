#!/bin/bash

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
