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

The design and implementation are currently in progress. The design is covered under this [design proposal PR](https://github.com/openebs/node-disk-manager/pull/1)



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fnode-disk-manager?ref=badge_large)
