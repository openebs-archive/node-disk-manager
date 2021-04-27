# node-disk-manager

[![Build Status](https://github.com/openebs/node-disk-manager/actions/workflows/build.yml/badge.svg)](https://github.com/openebs/node-disk-manager/actions/workflows/build.yml)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/ea8d7835d7224178af058d98e5dac117)](https://www.codacy.com/app/OpenEBS/node-disk-manager?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=openebs/node-disk-manager&amp;utm_campaign=Badge_Grade)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/node-disk-manager)](https://goreportcard.com/report/github.com/openebs/node-disk-manager)
[![codecov](https://codecov.io/gh/openebs/node-disk-manager/branch/master/graph/badge.svg)](https://codecov.io/gh/openebs/node-disk-manager)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![BCH compliance](https://bettercodehub.com/edge/badge/openebs/node-disk-manager?branch=master)](https://bettercodehub.com/results/openebs/node-disk-manager)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager?ref=badge_shield)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/1953/badge)](https://bestpractices.coreinfrastructure.org/projects/1953)

`node-disk-manager` (NDM) aims to make it easy to manage the disks attached to the node. It treats disks as resources that need to be monitored and managed just like other resources like CPU, Memory and Network. It contains a daemon which runs on each node, detects attached disks and loads them as BlockDevice objects (custom resource) into Kubernetes. 

While PVs are well suited for stateful workloads, the BlockDevice objects are aimed towards helping hyper-converged Storage Operators by providing abilities like:
- Easy to access inventory of block devices available across the Kubernetes Cluster.
- Predict failures on the blockdevices, to help with taking preventive actions.
- Allow for dynamically attaching/detaching blockdevices to a Storage Pod, without requiring a restart.

NDM has 2 main components:
- node-disk-manager daemonset, which runs on each node and is responsible for device detection.
- node-disk-operator deployment, which acts as an inventory of block devices in the cluster.

and 2 optional components:
- ndm-cluster-exporter deployment, which fetches block device object from etcd and exposes it as prometheus metrics.
- ndm-node-exporter daemonset, which runs on each node, queries the disk for details like SMART and expose it as prometheus metrics.

The design of the project is covered under this [design proposal](./docs/design.md)

## Project Status
Currently, the NDM project is in beta.

# Usage
A detailed usage documentation is maintained in the [wiki](https://github.com/openebs/node-disk-manager/wiki).

## Start Node Disk Manager
* Edit [ndm-operator.yaml](deploy/ndm-operator.yaml) to fit your environment: Set the `namespace`, `serviceAccount`, configure filters in the `node-disk-manager-config-map`.
* Switch to Cluster Admin context and create the components with `kubectl create -f ndm-operator.yaml`.
* This will install the daemon, operator and the exporters

## Using `kubectl` to fetch BlockDevice Information
* `kubectl get blockdevices` displays the blockdevices across the cluster, with `NODENAME` showing the node to which disk is attached,
  `CLAIMSTATE` showing whether the device is currently in use and `STATE` showing whether the device is connected to the node.
* `kubectl get blockdevices -o wide` displays the blockdevice along with the path at which the device is attached on the node.
* `kubectl get blockdevices <blockdevice-cr-name> -o yaml` displays all the details of the disk captured by `ndm` for given disk resource.

## Building, Testing and Pushing Image
Before building the image locally, you need to setup your development environment. The detailed instructions for setting up development environment, building and testing are available [here](./BUILD.md).

#### Push Image
By default Github Action pushes the docker image to `openebs/node-disk-manager`, with *ci* tags. 
You can push to your custom registry and modify the ndm-operator.yaml file for your testing. 

# License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager?ref=badge_large)

# Inspiration
* Thanks to [Daniel](https://github.com/dswarbrick) for setting up the go-based [SMART](https://github.com/dswarbrick/smart) library.
* Thanks to [Humble](https://github.com/humblec), [Jan](https://github.com/jsafrane) and other from the [Kubernetes Storage Community](https://github.com/kubernetes-incubator/external-storage/issues/736) for reviewing the approach and evaluating the use-case. 



