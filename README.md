# node-disk-manager

[![Build Status](https://travis-ci.org/openebs/node-disk-manager.svg?branch=master)](https://travis-ci.org/openebs/node-disk-manager)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/node-disk-manager)](https://goreportcard.com/report/github.com/openebs/node-disk-manager)
[![codecov](https://codecov.io/gh/openebs/node-disk-manager/branch/master/graph/badge.svg)](https://codecov.io/gh/openebs/node-disk-manager)
[![BCH compliance](https://bettercodehub.com/edge/badge/openebs/node-disk-manager?branch=master)](https://bettercodehub.com/results/openebs/node-disk-manager)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/openebs/node-disk-manager/blob/master/LICENSE)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager?ref=badge_shield)

node-disk-manager aims to make it easy to manage the disks attached to the node. It treats disks as resources that need to be monitored and managed just like other resources like CPU, Memory and Network. It is a daemon which runs on each node, detects attached disks and loads them as Disk objects (custom resource) into Kubernetes. 

While PVs are well suited for stateful workloads, the Disk objects are aimed towards helping hyper-converged Storage Operators by providing abilities like:
- Easy to access inventory of Disks available across the Kubernetes Cluster
- Predict failures on the Disks, to help with taking preventive actions
- Allow for dynamically attaching/detaching Disks to a Storage Pod, without requiring a restart

The design and implementation are currently in progress. The design is covered under this [design proposal](./docs/design.md)

# Usage
A detailed usage documentation is maintained in the [wiki](https://github.com/openebs/node-disk-manager/wiki)

## Start Node Disk Manager DaemonSet
* Edit [ndm-operator.yaml](./ndm-operator.yaml) to fit your environment: Set the `namespace`, `serviceAccount`, configure filters in the `node-disk-manager-config-map`.
* Switch to Cluster Admin context and create the DaemonSet with `kubectl create -f ndm-operator.yaml`

## Using `kubectl` to fetch Disk Information
* `kubectl get disks --show-labels` displays the disks across the cluster, with `kubernetes.io/hostname` showing the node to which disk is attached. 
* `kubectl get disks -l "kubernetes.io/hostname=<hostname>"` displays the disks attached to node with the provided hostname.
* `kubectl get disk <disk-cr-name> -o yaml` displays all the details of the disk captured by `ndm` for given disk resource

## Build Image
* `go get` or `git clone` node-disk-manager repo into `$GOPATH/src/github.com/openebs/`
with one of the below directions:
  * `cd $GOPATH/src/github.com/openebs && git clone git@github.com:openebs/node-disk-manager.git`
  * `cd $GOPATH/src/github.com/openebs && go get github.com/openebs/node-disk-manager`

* Setup build tools:
  * By default node-disk-manager enables fetching disk attributes using udev. This requires udev develop files. For Ubuntu, `libudev-dev` package should be installed.
  * `make bootstrap` installs the required Go tools

* run `make` in the top directory. It will:
  * Build the binary.
  * Build the docker image with the binary.

* Test your changes
  * `sudo -E env "PATH=$PATH" make test` execute the unit tests
  * `make integration-test` will launch minikube to run the tests. Make sure that minikube can be executed via `sudo -E minikube start --vm-driver=none`

## Push Image
By default travis pushes the docker image to `openebs/node-disk-manager-amd64`, with *ci* as well as commit tags. 
You can push to your custom registry and modify the ndm-operator.yaml file for your testing. 

# License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager?ref=badge_large)

# Inspiration
* Thanks to [Daniel](https://github.com/dswarbrick) for setting up the go-based [SMART](https://github.com/dswarbrick/smart) library.
* Thanks to [Humble](https://github.com/humblec), [Jan](https://github.com/jsafrane) and other from the [Kubernetes Storage Community](https://github.com/kubernetes-incubator/external-storage/issues/736) for reviewing the approach and evaluating the usecase. 



