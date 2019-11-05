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
