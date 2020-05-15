v0.5.0-RC2 / 2020-05-13
========================

  * add env to enable/disable CRD installation ([#421](https://github.com/openebs/node-disk-manager/pull/421), [@akhilerm](https://github.com/akhilerm))
  
v0.5.0-RC1 / 2020-05-08
========================

  * add support for blockdevice metrics using seachest ([#349](https://github.com/openebs/node-disk-manager/pull/349), [@akhilerm](https://github.com/akhilerm))
  * add ppc64le builds ([#374](https://github.com/openebs/node-disk-manager/pull/374), [@Pensu](https://github.com/Pensu))
  * add support for partitions and enable the new UUID algorithm for blockdevice UUID generation ([#386](https://github.com/openebs/node-disk-manager/pull/386), [@akhilerm](https://github.com/akhilerm))
  * add OpenEBS to the list of default excluded vendors ([#409](https://github.com/openebs/node-disk-manager/pull/409), [@akhilerm](https://github.com/akhilerm))
  * add new filter to validate BlockDevices and remove invalid entries ([#410](https://github.com/openebs/node-disk-manager/pull/410), [@akhilerm](https://github.com/akhilerm))
  * remove controller for cluster scoped disk resource ([#412](https://github.com/openebs/node-disk-manager/pull/412), [@akhilerm](https://github.com/akhilerm))
  * add finalizer on claimed BlockDevice resource to prevent accidental deletion ([#416](https://github.com/openebs/node-disk-manager/pull/416), [@akhilerm](https://github.com/akhilerm))

v0.4.9 / 2020-04-15
========================

  * add physical/logical block size, hardware sector size and drive type into BlockDevice resource
  ([#388](https://github.com/openebs/node-disk-manager/pull/388), [@akhilerm](https://github.com/akhilerm))
  * add new sysfs probe to fetch block device details from sysfs ([#375](https://github.com/openebs/node-disk-manager/pull/375), 
  [shovanmaity](https://github.com/shovanmaity))
  * enable persisting of labels and annotations on BlockDevice resource ([#394](https://github.com/openebs/node-disk-manager/pull/394), 
  [shovanmaity](https://github.com/shovanmaity))
  * add label selector to BlockDeviceClaim resource ([#397](https://github.com/openebs/node-disk-manager/pull/397), 
  [@akhilerm](https://github.com/akhilerm))
  * add support for `openebs.io/block-device-tag` label on BlockDevice ([#400](https://github.com/openebs/node-disk-manager/pull/400), 
  [@akhilerm](https://github.com/akhilerm))

v0.4.9-RC1 / 2020-04-07
========================

  * add physical/logical block size, hardware sector size and drive type into BlockDevice resource
  ([#388](https://github.com/openebs/node-disk-manager/pull/388), [@akhilerm](https://github.com/akhilerm))
  * add new sysfs probe to fetch block device details from sysfs ([#375](https://github.com/openebs/node-disk-manager/pull/375), 
  [shovanmaity](https://github.com/shovanmaity))
  * enable persisting of labels and annotations on BlockDevice resource ([#394](https://github.com/openebs/node-disk-manager/pull/394), 
  [shovanmaity](https://github.com/shovanmaity))
  * add label selector to BlockDeviceClaim resource ([#397](https://github.com/openebs/node-disk-manager/pull/397), 
  [@akhilerm](https://github.com/akhilerm))
  * add support for `openebs.io/block-device-tag` label on BlockDevice ([#400](https://github.com/openebs/node-disk-manager/pull/400), 
  [@akhilerm](https://github.com/akhilerm))

v0.4.8 / 2020-03-15
========================

  * enabled automatic builds of arm64 images ([#371](https://github.com/openebs/node-disk-manager/pull/371),
  [@akhilerm](https://github.com/akhilerm))

v0.4.8-RC1 / 2020-03-06
========================

  * enabled automatic builds of arm64 images ([#371](https://github.com/openebs/node-disk-manager/pull/371),
  [@akhilerm](https://github.com/akhilerm))

v0.4.7 / 2020-02-14
========================

  * added support to display blockdevice PATH in kubectl output ([#367](https://github.com/openebs/node-disk-manager/pull/367), 
  [@chandankumar4](https://github.com/chandankumar4))
  * customize location for NDM core files ([#362](https://github.com/openebs/node-disk-manager/pull/362), 
  [@akhilerm](https://github.com/akhilerm))
  
v0.4.7-RC1 / 2020-02-07
========================

  * added support to display blockdevice PATH in kubectl output ([#367](https://github.com/openebs/node-disk-manager/pull/367), 
  [@chandankumar4](https://github.com/chandankumar4))
  * customize location for NDM core files ([#362](https://github.com/openebs/node-disk-manager/pull/362), 
  [@akhilerm](https://github.com/akhilerm))
  
v0.4.6 / 2020-01-14
========================

  * added toleration to cleanup jobs ([#363](https://github.com/openebs/node-disk-manager/pull/363), 
  [@rahulchheda](https://github.com/rahulchheda))
  * disabled coredump in NDM Daemon by default ([#359](https://github.com/openebs/node-disk-manager/pull/359), 
  [@akhilerm](https://github.com/akhilerm))
  * disabled writing system wide core pattern ([#358](https://github.com/openebs/node-disk-manager/pull/358), 
  [@akhilerm](https://github.com/akhilerm)) 
  
v0.4.6-RC2 / 2020-01-11
========================

  * added toleration to cleanup jobs ([#363](https://github.com/openebs/node-disk-manager/pull/363), 
  [@rahulchheda](https://github.com/rahulchheda))

v0.4.6-RC1 / 2020-01-06
========================

  * disabled coredump in NDM Daemon by default ([#359](https://github.com/openebs/node-disk-manager/pull/359), 
  [@akhilerm](https://github.com/akhilerm))
  * disabled writing system wide core pattern ([#358](https://github.com/openebs/node-disk-manager/pull/358), 
  [@akhilerm](https://github.com/akhilerm)) 

v0.4.5 / 2019-12-13
========================

  * fixed security vulnerability in images used in cleanup pods ([#351](https://github.com/openebs/node-disk-manager/pull/351), 
  [@kmova](https://github.com/kmova))
  * added disk hierarchy information to the daemon logs ([#353](https://github.com/openebs/node-disk-manager/pull/353), 
  [@akhilerm](https://github.com/akhilerm))
  * ability to disable reconciliation for NDM resources ([#307](https://github.com/openebs/node-disk-manager/pull/307),
  [@akhilerm](https://github.com/akhilerm))

v0.4.5-RC2 / 2019-12-12
========================
   
  * ability to disable reconciliation for NDM resources ([#307](https://github.com/openebs/node-disk-manager/pull/307),
  [@akhilerm](https://github.com/akhilerm))

v0.4.5-RC1 / 2019-12-05
========================
  
  * fixed security vulnerability in images used in cleanup pods ([#351](https://github.com/openebs/node-disk-manager/pull/351), 
  [@kmova](https://github.com/kmova))
  * added disk hierarchy information to the daemon logs ([#353](https://github.com/openebs/node-disk-manager/pull/353), 
  [@akhilerm](https://github.com/akhilerm))

v0.4.4 / 2019-11-12
=======================

  * fix device detection for KVM based virtual machines
  * fix parent disk detection for nvme devices
  * refactor NDM cli to remove unused flags
  * add prometheus exporter for fetching metrics from etcd
  * replace glog with klog
  * refactor logs for easier parsing to send log based alerts


v0.4.4-RC2 / 2019-11-09
=======================

  * refactor logs for easier parsing to send log based alerts
  * fix device detection for KVM based virtual machines
   
v0.4.4-RC1 / 2019-11-05
=======================

  * add prometheus exporter for fetching metrics from etcd
  * replace glog with klog
  * fix parent disk detection for nvme devices
  * refactor NDM cli to remove unused flags

v0.4.3 / 2019-10-14
=======================

  * add support for building NDM on multiple platforms/architectures
  * support for arm64
  * refactored integration tests to remove dependency on minikube

v0.4.2 / 2019-09-09
=======================

  * add service account to cleanup job
  * support cancelling ongoing cleanup jobs
  * add filter to claim blockdevices based on nodename
  * fix os-disk filter to exclude empty disk paths

v0.4.1 / 2019-07-31
=======================

  * fix seachest holding on to open FDs
  * use controller-runtime signals for signal handling
  * automated installation of NDM CRDs from NDM operator
  * handle CRD upgrade in NDM operator
  * add extra check before removing finalizer on blockdeviceclaims
  * cleaned the NDM operator logs
  * change hostname to nodename and added support for nodename in blockdevice

v0.4.0 / 2019-06-21
=======================

  * introduce blockdevice resource for managing all blockdevices on the system
  * introduce blockdeviceclaim resource for claiming and unclaiming 
    blockdevices
  * introduce NDM operator for managing blockdeviceclaim
  * add scrub job to clean the blockdevice once it is unclaimed
  * add probe to get mount information of blockdevices
  * add integration test for disk attach, dynamic disk attach and disk
    detach operations

v0.3.5 / 2019-04-25
=======================

  * add support for NDM to run on device with SELinux 

v0.3.4 / 2019-04-09
=======================

  * fix NDM crash when udev probe failed to probe the disks. 

v0.3.3 / 2019-03-26
=======================

  * added GOTRACEBACK to print stack trace 

v0.3.2 / 2019-03-01
=======================

  * support for sparse file size given in exponential format 

v0.3.1 / 2019-02-26
=======================

  * fix NDM restart when disk is having less than 3 partitions. 

v0.3.0 / 2019-02-22
=======================

  * enable core dump for NDM
  * add seachest probe to get additional disk details for physical disks
  * added partition and filesystem information
  * add support for unmanaged disks
  * fix crash issue when NDM is run in unprivileged mode. Fallback to
    limited feature set instead of crashing
  * added integration tests for path-filter

v0.2.0 / 2018-10-25
=======================

  * fix readDeviceCapacity method to handle disks > 2TB
  * add probe to determine capacity when udev doesn't support size
  * add configurable filter to determine os disk via config map
  * refactor config map to use yaml format
  * refactor to push docker images to quay repo along with docker 
  * refactor to fix lint warnings in several files
  * add selectable github issue templates

v0.1.0-RC3 / 2018-09-01
==================

  * support configuring filters based on disk path patterns. Example:
    - Exclude disks where path includes `loop`
  * support generating uuid with path and hostname for below disk types
    where wwn,serial,model,vendor are either not present or missing:
    - AWS Ephemeral SSDs 
    - GKE Ephemeral SSDs 
    - VMWare Virtual Disks 

v0.1.0-RC2 / 2018-08-22
=======================

  * include support for creating sparse file

v0.1.0-RC1 / 2018-08-08
=======================

  * set Disk status as unknown when ndm pod is being shutdown
  * add NDM ConfigMap to customize filters and probes
  * filter disks based on the vendor type
  * filter os disk while creating disk cr
  * support probing via mod pages to fetch basic disk attributes
  * order devlinks to place by-id links as the first link in Disk DR
  * detect and process disk add/remove events; create or update status
  * add devlinks to disk cr
  * add hostname as a label to Disk CR
  * support probing the disks using udev and remove lsblk based discovery
  * create ndm-operator.yaml to install Disk DRD and NDM as Daemonset
  * auto-generate client code to access Disk CR
  * add a dockerfile and .travis.yml to build node-disk-manager
  * use kubernetes go-client - 6.0.0
  * use hash of wwn,serial,model,vendor to generate uuid.
  * discover disks via the lsblk system command.
