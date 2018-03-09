# Introduction 

Kubernetes adoption for stateful workloads is increasing, along with the need for having hyper-converged storage solutions that can make use of the disks directly attached to the Kubernetes nodes. The current options to manage the disks are either:
- Use Local PVs that directly expose a disk as a PV to the applications. These are intended for applications that can take care of replication or data protection.
- Use hyper-converged storage solutions or Container Attached Storage solutions like OpenEBS that consume the underlying disks and provide PVs to workloads that also include capabilities like replication and data-protection (and many more standard storage features). 

The Container Attached Storage solutions - provide storage for containers (application workloads) using containers (storage pods).

One of the use-cases for the Local PVs was to provide a way for these storage engines (or storage pods) to consume Local PVs. However, the current scope of the Local PVs is not intended to implement the following additional requirements from the storage pods: 
- Need to know about the different types of disks (SSDs, NVMe devices, etc.,) available in the Cluster
- Need to be aware of internal layout in terms of NUMA to help assigning the right cores to the storage containers for faster dedicated access (on the lines of using DPDK/SPDK).
- Need to determine the physical location of the disks in the cluster in terms of Node, Rack, Data Center/Zone
- Need to associate the required storage to OpenEBS Storage Pods, taking into consideration the capacity, qos and availability requirements 
- Need immediate notifications on the faults on the storage disks and ability to perform online corrections

One of the primary blocker to using the local PV by Storage Pods is the hard-requirement of a need to restart a pod, whenever new PVs are attached or detached to a Pod.

This has led to most of the storage solution provider to build the capabilities into their solution, which could be abstracted out and used by multiple storage vendors. 

Refer:
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/local-storage-pv.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/local-storage-pv.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-topology-scheduling.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-metrics.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/raw-block-pv.md


## Proposal Overview

This proposal introduces a new component called _*node-disk-manager*_ that will automate the management of the disks attached to the node.

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


## Proposed Solution

No changes to the current approach of accessing the disks by the storage pods as followed by Storage Solutions like OpenEBS, PortWorx, StorageOS, GlusterFS, Rook/Ceph. The storage pods will continue to mount "/dev" directories and get access to the actual disks.

The change is primarily in the orchestration(aka orchestration) of how disks attached to the node are:
- discovered
- provisioned
- monitored
- assigned to Storage Pools - Aggregate of Disks 

The new work flow with _*node-disk-manager*_ would be:

- node-disk-manager can be configured to run on all node or select nodes (designated as Storage Nodes) 

- The operator that installs the node-disk-manager will also insert/create CRDs for Disk Object and StoragePool.

- On start, the node-disk-manager will discover the disks attached to the node where it is running and will create Disk Object CRs if they were not already present. 

- A Disk Object CR will contain information like:
  * type
  * resource identifier
  * physical location - node, rack, zone

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

