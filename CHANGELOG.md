
n.n.n / 2018-09-01
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
  * add devlinks to disk cr
  * add hostname as a label to Disk CR
  * support probing the disks using udev and remove lsblk based discovery
  * create ndm-operator.yaml to install Disk DRD and NDM as Daemonset
  * auto-generate client code to access Disk CR
  * add a dockerfile and .travis.yml to build node-disk-manager
  * use kubernetes go-client - 6.0.0
  * use hash of wwn,serial,model,vendor to generate uuid.
  * discover disks via the lsblk system command.
