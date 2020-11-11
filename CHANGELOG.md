v1.0.0-RC1 / 2020-11-11
========================
* add controller options to device list command, fixed sysfs probe processing empty devices ([#504](https://github.com/openebs/node-disk-manager/pull/504),[@akhilerm](https://github.com/akhilerm))
* chore(build) Updating dockerfile(s) with buildx built-in ARGs ([#503](https://github.com/openebs/node-disk-manager/pull/503),[@xunholy](https://github.com/xunholy))
* add support for device-mapper(dm) devices. ([#495](https://github.com/openebs/node-disk-manager/pull/495),[@akhilerm](https://github.com/akhilerm))
* restrict claiming of blockdevices with empty block-device-tag value ([#500](https://github.com/openebs/node-disk-manager/pull/500),[@ajeetrai707](https://github.com/ajeetrai707))



v0.9.0 / 2020-10-14
========================
* (fix) Support excluding multiple OS disk paths ([#481](https://github.com/openebs/node-disk-manager/pull/481),[@rahulchheda](https://github.com/rahulchheda))
* fix a bug where partition table was written on disk with filesystem, resulting in data loss ([#496](https://github.com/openebs/node-disk-manager/pull/496),[@akhilerm](https://github.com/akhilerm))
* add partition name into NDM created partitions ([#494](https://github.com/openebs/node-disk-manager/pull/494),[@avats-dev](https://github.com/avats-dev))
* fix(mount): detect real device when using /dev/root ([#492](https://github.com/openebs/node-disk-manager/pull/492),[@zlymeda](https://github.com/zlymeda))


v0.9.0-RC1 / 2020-10-08
========================
* (fix) Support excluding multiple OS disk paths ([#481](https://github.com/openebs/node-disk-manager/pull/481),[@rahulchheda](https://github.com/rahulchheda))
* fix a bug where partition table was written on disk with filesystem, resulting in data loss ([#496](https://github.com/openebs/node-disk-manager/pull/496),[@akhilerm](https://github.com/akhilerm))
* add partition name into NDM created partitions ([#494](https://github.com/openebs/node-disk-manager/pull/494),[@avats-dev](https://github.com/avats-dev))
* fix(mount): detect real device when using /dev/root ([#492](https://github.com/openebs/node-disk-manager/pull/492),[@zlymeda](https://github.com/zlymeda))



v0.8.1 / 2020-09-15
========================
* add support to add custom tag to blockdevices based on config ([#475](https://github.com/openebs/node-disk-manager/pull/475),[@akhilerm](https://github.com/akhilerm))
* fix a bug where NDM operator crashes if a claimed BD is manually deleted ([#479](https://github.com/openebs/node-disk-manager/pull/479),[@akhilerm](https://github.com/akhilerm))
* update go version to 1.14.7 ([#476](https://github.com/openebs/node-disk-manager/pull/476),[@akhilerm](https://github.com/akhilerm))
* add additional check to exclude blockdevices with tag while manual claiming ([#404](https://github.com/openebs/node-disk-manager/pull/404),[@akhilerm](https://github.com/akhilerm))


v0.8.1-RC1 / 2020-09-10
========================
* add support to add custom tag to blockdevices based on config ([#475](https://github.com/openebs/node-disk-manager/pull/475),[@akhilerm](https://github.com/akhilerm))
* fix a bug where NDM operator crashes if a claimed BD is manually deleted ([#479](https://github.com/openebs/node-disk-manager/pull/479),[@akhilerm](https://github.com/akhilerm))
* update go version to 1.14.7 ([#476](https://github.com/openebs/node-disk-manager/pull/476),[@akhilerm](https://github.com/akhilerm))
* add additional check to exclude blockdevices with tag while manual claiming ([#404](https://github.com/openebs/node-disk-manager/pull/404),[@akhilerm](https://github.com/akhilerm))



v0.8.0 / 2020-08-14
========================
* Upgrade go version to 1.14 ([#459](https://github.com/openebs/node-disk-manager/pull/459),[@harshthakur9030](https://github.com/harshthakur9030))
* Remove dependency on gox and instead use native go build. ([#456](https://github.com/openebs/node-disk-manager/pull/456),[@harshthakur9030](https://github.com/harshthakur9030))
* Make udev scan operation thread safe. ([#455](https://github.com/openebs/node-disk-manager/pull/455),[@harshthakur9030](https://github.com/harshthakur9030))
* remove v prefix from all image tags ([#467](https://github.com/openebs/node-disk-manager/pull/467),[@akhilerm](https://github.com/akhilerm))
* automate migration of blockdevices from legacy UUID to GPT Based UUID ([#442](https://github.com/openebs/node-disk-manager/pull/442),[@akhilerm](https://github.com/akhilerm))
* API Service to provide additional functionality ([#433](https://github.com/openebs/node-disk-manager/pull/433),[@harshthakur9030](https://github.com/harshthakur9030))
* fix running cleanup job for sparse blockdevices. ([#463](https://github.com/openebs/node-disk-manager/pull/463),[@akhilerm](https://github.com/akhilerm))
* update the project dependencies (k8s: 1.17.4, controller-runtime: 0.5.2, operator-sdk: 0.17.0) ([#365](https://github.com/openebs/node-disk-manager/pull/365),[@akhilerm](https://github.com/akhilerm))
* disable metrics server of controller runtime by default. ([#473](https://github.com/openebs/node-disk-manager/pull/473),[@akhilerm](https://github.com/akhilerm))


v0.8.0-RC2 / 2020-08-12
========================
* disable metrics server of controller runtime by default. ([#473](https://github.com/openebs/node-disk-manager/pull/473),[@akhilerm](https://github.com/akhilerm))


v0.8.0-RC1 / 2020-08-10
========================
* Upgrade go version to 1.14 ([#459](https://github.com/openebs/node-disk-manager/pull/459),[@harshthakur9030](https://github.com/harshthakur9030))
* Remove dependency on gox and instead use native go build. ([#456](https://github.com/openebs/node-disk-manager/pull/456),[@harshthakur9030](https://github.com/harshthakur9030))
* Make udev scan operation thread safe. ([#455](https://github.com/openebs/node-disk-manager/pull/455),[@harshthakur9030](https://github.com/harshthakur9030))
* remove v prefix from all image tags ([#467](https://github.com/openebs/node-disk-manager/pull/467),[@akhilerm](https://github.com/akhilerm))
* automate migration of blockdevices from legacy UUID to GPT Based UUID ([#442](https://github.com/openebs/node-disk-manager/pull/442),[@akhilerm](https://github.com/akhilerm))
* API Service to provide additional functionality ([#433](https://github.com/openebs/node-disk-manager/pull/433),[@harshthakur9030](https://github.com/harshthakur9030))
* fix running cleanup job for sparse blockdevices. ([#463](https://github.com/openebs/node-disk-manager/pull/463),[@akhilerm](https://github.com/akhilerm))
* update the project dependencies (k8s: 1.17.4, controller-runtime: 0.5.2, operator-sdk: 0.17.0) ([#365](https://github.com/openebs/node-disk-manager/pull/365),[@akhilerm](https://github.com/akhilerm))



v0.7.0 / 2020-07-14
========================
* fix wiping released blockdevices with partitions ([#445](https://github.com/openebs/node-disk-manager/pull/445), [@akhilerm](https://github.com/akhilerm))
* fix bug of having an open file descriptor in NDM causing applications to receive resource busy error. ([#450](https://github.com/openebs/node-disk-manager/pull/450), [@akhilerm](https://github.com/akhilerm))
* Adding support to build multi-arch docker images. ([#428](https://github.com/openebs/node-disk-manager/pull/428), [@xunholy](https://github.com/xunholy))
* deprecate invalid capacity request phase from block device claim ([#443](https://github.com/openebs/node-disk-manager/pull/443), [@akhilerm](https://github.com/akhilerm))


v0.7.0-RC1 / 2020-07-09
========================
* fix wiping released blockdevices with partitions ([#445](https://github.com/openebs/node-disk-manager/pull/445), [@akhilerm](https://github.com/akhilerm))
* fix bug of having an open file descriptor in NDM causing applications to receive resource busy error. ([#450](https://github.com/openebs/node-disk-manager/pull/450), [@akhilerm](https://github.com/akhilerm))
* Adding support to build multi-arch docker images. ([#428](https://github.com/openebs/node-disk-manager/pull/428), [@xunholy](https://github.com/xunholy))
* deprecate invalid capacity request phase from block device claim ([#443](https://github.com/openebs/node-disk-manager/pull/443), [@akhilerm](https://github.com/akhilerm))


v0.6.0 / 2020-06-13
========================
* make feature gates independent of daemon controller ([#426](https://github.com/openebs/node-disk-manager/pull/426), [@akhilerm](https://github.com/akhilerm))
* remove all disk resources and disk CRD as part of installation ([#427](https://github.com/openebs/node-disk-manager/pull/427), [@akhilerm](https://github.com/akhilerm))
* add new discovery probe (called used-by-probe) to detect if devices are used by K8s Local PV, ZFS-LocalPV, Mayastor and cStor ([#430](https://github.com/openebs/node-disk-manager/pull/430), [@akhilerm](https://github.com/akhilerm))
* migrate project to use go modules ([#434](https://github.com/openebs/node-disk-manager/pull/434), [@harshthakur9030](https://github.com/harshthakur9030))
* Adding filesystem info column in output of kubectl get bd -o wide ([#435](https://github.com/openebs/node-disk-manager/pull/435), [@harshthakur9030](https://github.com/harshthakur9030))


v0.6.0-RC2 / 2020-06-12
========================
* Adding filesystem info column in output of kubectl get bd -o wide ([#435](https://github.com/openebs/node-disk-manager/pull/435), [@harshthakur9030](https://github.com/harshthakur9030))


v0.6.0-RC1 / 2020-06-09
========================
* make feature gates independent of daemon controller ([#426](https://github.com/openebs/node-disk-manager/pull/426), [@akhilerm](https://github.com/akhilerm))
* remove all disk resources and disk CRD as part of installation ([#427](https://github.com/openebs/node-disk-manager/pull/427), [@akhilerm](https://github.com/akhilerm))
* add new discovery probe (called used-by-probe) to detect if devices are used by K8s Local PV, ZFS-LocalPV, Mayastor and cStor ([#430](https://github.com/openebs/node-disk-manager/pull/430), [@akhilerm](https://github.com/akhilerm))
* migrate project to use go modules ([#434](https://github.com/openebs/node-disk-manager/pull/434), [@harshthakur9030](https://github.com/harshthakur9030))


v0.5.0 / 2020-05-15
========================

  * add support for blockdevice metrics using seachest ([#349](https://github.com/openebs/node-disk-manager/pull/349), [@akhilerm](https://github.com/akhilerm))
  * add ppc64le builds ([#374](https://github.com/openebs/node-disk-manager/pull/374), [@Pensu](https://github.com/Pensu))
  * add support for partitions and enable the new UUID algorithm for blockdevice UUID generation ([#386](https://github.com/openebs/node-disk-manager/pull/386), [@akhilerm](https://github.com/akhilerm))
  * add OpenEBS to the list of default excluded vendors ([#409](https://github.com/openebs/node-disk-manager/pull/409), [@akhilerm](https://github.com/akhilerm))
  * add new filter to validate BlockDevices and remove invalid entries ([#410](https://github.com/openebs/node-disk-manager/pull/410), [@akhilerm](https://github.com/akhilerm))
  * remove controller for cluster scoped disk resource ([#412](https://github.com/openebs/node-disk-manager/pull/412), [@akhilerm](https://github.com/akhilerm))
  * add finalizer on claimed BlockDevice resource to prevent accidental deletion ([#416](https://github.com/openebs/node-disk-manager/pull/416), [@akhilerm](https://github.com/akhilerm))
  * add env to enable/disable CRD installation ([#421](https://github.com/openebs/node-disk-manager/pull/421), [@akhilerm](https://github.com/akhilerm))

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
