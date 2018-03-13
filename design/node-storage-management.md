# Introduction 

This document presents use-cases, high level design and workflow changes required for using disks attached to the Kubernetes Nodes by Container Attached Storages like OpenEBS. This design includes introducing new add-on component called - node-disk-manager, as well as extending some existing components where applicable like node-problem-detector, node-feature-discovery, etc. 

## Prerequisites

Design Proposals and components that are related to local disk management:
- https://github.com/kubernetes-incubator/node-feature-discovery
- https://github.com/kubernetes/node-problem-detector
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/local-storage-pv.md
- https://github.com/kubernetes-incubator/external-storage/tree/master/local-volume
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-topology-scheduling.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-metrics.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/raw-block-pv.md

## Goals

- Provide the ability to treat disks attached to the node as k8s objects and allow storage capabilities to be built on top of these objects before being used by the pods
- Provide a generic way to read the properties of disks, so that the applications above need not know the type of disk underneath
- Be able to pool the disks to augment the capacity or performance

## Background

Kubernetes adoption for stateful workloads is increasing, along with the need for having hyper-converged solutions that can make use of the disks directly attached to the Kubernetes nodes. The current options to manage the disks are either:
- Use Local PVs that directly expose a disk as a PV to the applications. These are intended for applications that can take care of replication or data protection.
- Use hyper-converged storage solutions or Container Attached Storage solutions like OpenEBS that consume the underlying disks and provide PVs to workloads that also include capabilities like replication and data-protection (and many more standard storage features). 

The Container Attached Storage solutions - provide storage for containers (application workloads) using containers (storage pods).

One of the use-cases for the Local PVs was to provide a way for these container attached storage engines (or storage pods) to consume Local PVs. The details of the proposed workflow are at: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/local-storage-overview.md#bob-manages-a-distributed-filesystem-which-needs-access-to-all-available-storage-on-each-node

The primary blocker for using local PV by Storage Pods is the hard-requirement of the need to restart a pod, whenever new PVs are attached or detached to a Pod (https://github.com/kubernetes/kubernetes/issues/58569). In addition, there is a need for an operator that: 
- has to know about the different types of disks (SSDs, NVMe devices, etc.,) available in the Cluster
- has to be aware of internal layout in terms of NUMA to help assigning the right cores to the storage containers for faster dedicated access (on the lines of using DPDK/SPDK).
- has to determine the physical location of the disks in the cluster in terms of Node, Rack, Data Center/Zone
- has to associate the required storage to OpenEBS Storage Pods, taking into consideration the capacity, qos and availability requirements 
- has to immediately send notifications on the faults on the storage disks and ability to perform online corrections

**node-disk-manager** that is proposed in this design is to lay the foundation for providing such operator as an Kubernetes add-on component, which can be further extended/customized to suite the specific container attached storage solution.

## Proposal Overview

This proposal introduces a new component called _*node-disk-manager*_ that will automate the management of the disks attached to the node. There will be new Custom Resources added to Kubernetes to represent the underlying storage infrastructure like - Disks and StoragePools and their associated Claims - DiskClaim, StoragePoolClaim. 

The disks could be of any type ranging from local disks (ssds or spinning media), NVMe/PCI based or disks coming from external SAN/NAS. A new Kubernetes Custom Resource called - Disk Object will be added that will represent the disks attached to the node.

The node-disk-manager can also help in creating Storage Pools using one or more Disk Objects. The Storage Pools can be of varying types starting from a plain ext4 mounts, lvm, zfs to ceph, glusterfs, cstor pools. 

node-disk-manager is intended to:
- be extensible in terms of adding support for different kinds of Disk Objects and Storage Pools. 
- act as a bridge to augment the functionality already provided by the kubelet or other infrastructure containers running on the node. For example, node-bot will have to interface with node-problem-detector or heapster etc., 
- be considered as a Infrastructure Pod (or a daemonset) that runs on each of the Kubernetes node or the functionality could be embedded into other Infrastructure Pods, similar to node-problem-detector. 


## Alternate Solutions

### Using Local PV

Within the Kubernetes, the closest construct that is available to represent a disk is a Local Persistent Volume. A typical flow for using the Local PVs by the Storage Pods could be as follows:
- Local PV Provisioner is used to create Local PVs representing the disks. The disks that have be used by the Local PV Provisioner have to be mounted into a directory (like /mnt/localdisks) and this location should be passed using a Config Map to the Local PV Provisioner. 

- Modify/extend the provisioner to add the following additional information:
  * PVs will be associated with node,rack information that will be obtained by the labels in the nodes. (TODO Kiran - Find the information on how nodes are associated with rack/zone information)
  * PVs will get additional information regarding the type of the disk, iops that is capable of etc., 

- Storage Solution Operators (say OpenEBS cStor Pool Provisioner or GlusterFS Provisioner) could then:
  * query for the PVs using `kubectl get pv` (filtered with local-provisioner type) to get information on all the disks in the Cluster.
  * create the storage pods by passing/binding them with the Local PVs. 

- Storage Solution Operators (say OpenEBS Operator or GlusterFS Operator) will have to monitor for faults on the PVs and send notification to the respective storage pods to take corrective actions. The monitoring includes information like:
  * gathering smart statistics of the disk
  * probing for mode-sense data on disks or listening for events generated by the underlying disk

However with this Local PV approach, any of the following operations will require a storage pod restart (Issue https://github.com/kubernetes/kubernetes/issues/58569 ):
  * new PVs have to attached to expand the storage available
  * Remove or replace a failed PV

A storage pod restarts can be expensive in terms of requiring data rebuilding and also introducing additional scenarios for failure combinations that could sometimes result in applications seeing high latencies or sometimes causing the volumes to get into read-only. 


### Using node-disk-manager

The workflow with _*node-disk-manager*_ would be to integrate into current approach of container attached storage solutions where - disks are accessed by mounting a single volume "/dev", while the _*node-disk-manager*_ will provide a generic way to:
* manage (discover, provision and monitor) disks
* manage storage pools (aggregate of disks)

- node-disk-manager can be configured to run on all node or select nodes (designated as Storage Nodes) 

- The operator that installs the node-disk-manager will also insert/create CRDs for Disk Object and StoragePool.

- On start, the node-disk-manager will discover the disks attached to the node where it is running and will create Disk Object CRs if they were not already present. 

- A Disk object CR will contain information like:
  * type
  * resource identifier
  * physical location - node, rack, zone
  ```
  apiVersion: openebs.io/v1
  kind: Disk
  metadata:
    labels:
      kubernetes.io/hostname: gke-openebs-kmova-default-pool-044afcb8-bmc0
    name: disk-3194ccbe-268f-11e8-b467-0ed5f89f718b
  spec:
    capacity:
      storage: 25G
    details:
      model: PersistentDisk
      serial: disk-node-bmc0
      vendor: Google
    path: /dev/sdb
  ```

- node-disk-manager can also be made to auto-provision disks by using a DiskClaim Object similar to (PVC that uses dynamic provisioners to create a PV). As an example, if the node-disk-manager is running on a Kubernetes Cluster in AWS, it can initiate a request to the underlying AWS to create a EBS and attach to the node. This functionality will be access controlled.

- The User or storage-operators( like OpenEBS provisioner) can use the kube-api/kubectl (like `kubectl get disks`) to get information on all the disks in the Cluster.

- The node-disk-manager, can be used to create StoragePools by taking one or more Disk Objects and putting a storage layer on top of it like ext4, lvm, zfs or spin-up containers like OpenEBS cStor Pool or Ceph OSD (Confirm this), etc., The StoragePool creation can be triggered via an StoragePoolClaim that the user can feed in via kubectl. Examples:
  * StoragePoolClaim can request for formatting a disk with ext4 and making it available at a certain path. 
  * StoragePoolClaim can request for a creation of a cStorPool - which is achieved by creating a cstor pod on top of the disks. The disks to be used by the cstor pod will be passed as configuration parameters, while mounting only "/dev" into cstor pod. To create a cstor pod, the node-disk-manager can call the api of openebs-provisioner. 

- node-disk-manager will monitor for faults on the Disk Objects and send notification to the subscribed listeners (say the openebs-provisioner if the fault was detected on the disk object using by OpenEBS cStor Pool). The monitoring includes information like:
  * gathering smart statistics of the disk
  * probing for mode-sense data on disks or listening for events generated by the underlying disk (could use integration into node-problem-detector here). 

- node-disk-manager can be configured to monitor for the StoragePools and take some corrective actions - like monitoring the cStorPool for data usage and automatically add new Disk Objects to the StoragePool or when there is a fault in the disk, replace it with another spare disk, etc.,.

Challenges:
- The underlying disks used/allocated by Storage Pools should not be assigned by Kubernetes (say via Local Provisioner) to different workloads/pods. Possibly this needs to be addressed by making sure that node-disk-manager doesn't use the disks that are assigned to the directories(like /mnt/localdisks) used by the Local PV Provisioner. 

## Detailed Design

### Installation/Setup

Following CRDs will be installed:
- [Disk](../crd/disk-crd.yaml)
- [DiskClaim](../crd/diskclaim-crd.yaml)
- [StoragePool](../crd/storagepool-crd.yaml)
- [StoragePoolClaim](../crd/storagepoolclaim-crd.yaml) 

_*node-disk-manager(ndm)*_ will be deployed as DaemonSet on all the storage nodes. The _**ndm**_ will discover all the disks attached to the node where it is running and will create corresponding Disk objects. `kubectl get disks` can be used to list the Disk objects. For example, if two of the nodes in the GKE Cluster had a 25G GPD attached to them, the output for `kubectl get disks -o yaml` is as follows: 
```
apiVersion: v1
items:
- apiVersion: openebs.io/v1
  kind: Disk
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"openebs.io/v1","kind":"Disk","metadata":{"annotations":{},"labels":{"kubernetes.io/hostname":"gke-openebs-kmova-default-pool-044afcb8-bmc0"},"name":"disk-3194ccbe-268f-11e8-b467-0ed5f89f718b","namespace":""},"spec":{"capacity":{"storage":"25G"},"details":{"model":"PersistentDisk","serial":"disk-node-bmc0","vendor":"Google"},"path":"/dev/sdb"}}
    clusterName: ""
    creationTimestamp: 2018-03-13T07:24:07Z
    labels:
      kubernetes.io/hostname: gke-openebs-kmova-default-pool-044afcb8-bmc0
    name: disk-3194ccbe-268f-11e8-b467-0ed5f89f718b
    namespace: ""
    resourceVersion: "12616"
    selfLink: /apis/openebs.io/v1/disk-3194ccbe-268f-11e8-b467-0ed5f89f718b
    uid: 841581eb-268f-11e8-9de2-42010aa00050
  spec:
    capacity:
      storage: 25G
    details:
      model: PersistentDisk
      serial: disk-node-bmc0
      vendor: Google
    path: /dev/sdb
- apiVersion: openebs.io/v1
  kind: Disk
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"openebs.io/v1","kind":"Disk","metadata":{"annotations":{},"labels":{"kubernetes.io/hostname":"gke-openebs-kmova-default-pool-044afcb8-lxh4"},"name":"disk-c9279540-a74a-49e2-833c-e46edd3db74c","namespace":""},"spec":{"capacity":{"storage":"25G"},"details":{"model":"PersistentDisk","serial":"disk-node-lxh4","vendor":"Google"},"path":"/dev/sdb"}}
    clusterName: ""
    creationTimestamp: 2018-03-13T07:24:16Z
    labels:
      kubernetes.io/hostname: gke-openebs-kmova-default-pool-044afcb8-lxh4
    name: disk-c9279540-a74a-49e2-833c-e46edd3db74c
    namespace: ""
    resourceVersion: "12625"
    selfLink: /apis/openebs.io/v1/disk-c9279540-a74a-49e2-833c-e46edd3db74c
    uid: 897b7729-268f-11e8-9de2-42010aa00050
  spec:
    capacity:
      storage: 25G
    details:
      model: PersistentDisk
      serial: disk-node-lxh4
      vendor: Google
    path: /dev/sdb
```



## Feature Implementation Plan

Following is a high level summary of the feature implementation.

### Phase 1
- Discover the attached disk resources and push them into etcd as Disk custom resources
- Support for creating a StoragePool (type=ext4) in response to StoragePoolClaim 
- Support for monitoring disk latencies via mode-sense probes and raising Events
- Support for monitoring disk errors via node-problem-detector and notifying Events
- Support for capacity usage metrics per StoragePool that can be consumed by Prometheus
- Support for iostats metrics per Disk that can be consumed by Prometheus


### Phase 2
- Support for creating a StoragePool (type=cstor) in response to StoragePoolClaim 
- Support for notifying events to API endpoints

### Future
- Dynamic Provisioning of disks from external storage in response to DiskClaim custom resource from User.
- Consider disks that move from one node to another or can be attached to multiple nodes (as is the case with external storage disks)

