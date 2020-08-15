# API service user guide

## What does this API do?
The API service exposes some functionality of NDM to the user which can help in gaining additional information about block devices and troubleshooting. It is exposed as a gRPC service.

## Currently available services:

- Version : Gives the version and git commit of NDM which is running

- Node name : Gives the name of the node it is running on

- List Block Devices : Similar to `lsblk`, it lists all the block devices it finds and relationships among them.

- Block Device Details:  Given the block device path, it can give [S.M.A.R.T](https://en.wikipedia.org/wiki/S.M.A.R.T.) stats of that block device. 

- iSCSI status : Since most of the data engines supported by OpenEBS require iSCSI, it's status can be checked from here. 

- Hugepages : Mayastor requires hugepages to be set on the node. Hence, you can use the API to set hugepages and check the number of hugepages available on the node. 
**Note : Running the SET method does NOT guarantee that the hugepages will be set, it's recommended to verify by using the GET method.**

- Rescan : Instead of restarting the NDM pod when something seems wrong, Rescan can be run to sync NDM's local state with etcd. NDM primarily relies on [udev](https://opensource.com/article/18/11/udev) events and rarely, the events can be missed by NDM.
 For instance, a disk could be attached to the node but you don't see it in output of `kubectl get bd` . In such a case, running a rescan would help. Rescan also helps when there's a change in capacity of the block device or when a filesystem is mounted. All the devices will be scanned and latest information will be updated to etcd.

## How to use it?
CLI for accessing the service is not completely implemented. A client like [grpcurl](https://github.com/fullstorydev/grpcurl) can be used currently to access the gRPC service.
