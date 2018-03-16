# node-disk-manager

node-disk-manager aims to make it easy to manage the disks attached to the node. It treats disks as resources that need to be monitored and managed just like other resources like CPU, Memory and Network. It is a daemon which runs on each node, detects attached disks and loads them as Disk objects (custom resource) into Kubernetes. 

While PVs are well suited for stateful workloads, the Disk objects are aimed towards helping hyper-converged Storage Operators by providing abilities like:
- Easy to access inventory of Disks available across the Kubernetes Cluster
- Predict failures on the Disks, to help with taking preventive actions
- Allow for dynamically attaching/detaching Disks to a Storage Pod, without requiring a restart

The design and implementation are currently in progress. The design is covered under this [design proposal PR](https://github.com/openebs/node-disk-manager/pull/1)

