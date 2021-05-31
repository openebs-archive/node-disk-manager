# Mount-point change detection

## Introduction

This document presents the design of the mount-point change detection system in NDM.
The goal of mount-point change detection is to detect the changes in the mount-points
and the filesystem on the existing blockdevices discovered by NDM and trigger appropriate
action to update the blockdevice CRs.

## Design

The mount-points and filesystem info of a block device are found by the probe - *mountprobe*.
This probe uses the mounts file, which the Linux kernel provides in the *procfs* pseudo
filesystem. Reading the mounts file provides the status of all the mounted
filesystems (sample output below).

``` text
rootfs / rootfs rw 0 0
/dev/root / ext3 rw 0 0
/proc /proc proc rw 0 0 usbdevfs
/proc/bus/usb usbdevfs rw 0 0
/dev/sda1 /boot ext3 rw 0 0 none
/dev/pts devpts rw 0 0
/dev/sda4 /home ext3 rw 0 0 none
/dev/shm tmpfs rw 0 0 none
/proc/sys/fs/binfmt_misc binfmt_misc rw 0 0
```

Whenever a block device is (un)mounted or the fs type changes, the changes are reflected in the mounts file. This proposal introduces a change to the existing _**mount-probe**_ in NDM. Similar to how _udev-probe_ listens to udev events by starting a loop in its `Start()`, a loop is started by _mount-probe_ that watches for changes in the mounts file and triggers updation when a change is detected. The *epoll* API is used to watch the mounts file for changes. The *epoll* API is provided by the Linux kernel for the userspace programs to monitor file descriptors and get notifications about I/O events happening on them. Whenever the mounts file changes, the events `EPOLLPRI` and `EPOLLERR` are emitted. This behaviour has been documented [here](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=31b07093c44a7a442394d44423e21d783f5523b8) (additional links - [\[1\]](https://lkml.org/lkml/2006/2/22/169), [\[2\]](http://lkml.iu.edu/hypermail/linux/kernel/1012.1/02246.html)).

An important thing to note here is that while this approach can tell that the mounts file has changed, we cannot determine what changed. We cannot determine which block device is the change related to or whether the change is even related to a block device. Also, filesystem changes can be detected only once a device is mounted. Formatting filesystem on an unmounted device will not result in any updates.

The listen loop continously waits for the `EPOLLPRI` and `EPOLLERR` events. On receving such an event, a `EventMessage` is generated and pushed to `udevevent.UdevEventMessageChannel` channel. The message
contains information about what blockdevices to check (all existing blockdevices in this case). The message also
has additional information regarding what probes are to be run and specifies that only _mount-probe_ needs to be run for the event. This is done since _mount-probe_ alone can fetch the new mounts and fs data for the blockdevices. Running the probes selectively helps us optimize the updation process.
The message is then received by the loop in `udevProbe.listen()` and sent further down to the `ProbeEvent` update handler.

&nbsp;

``` text
+---------------------+                                                                    +------------------------+
|                     |                                                                    |                        |
|      Epoll Api      |                                                                    |       ProbeEvent       |
|                     |                                                                    |                        |
+----------+----------+                                                                    |     Update Handler     |
           |                                                                               |                        |
           |                                                                               +------------^-----------+
           |                                                                                            |
           |   (EPOLLPRI & EPOLLERR)                                                                    |
           |                                                                                            |
           |                                                                               EventMessage |
           |                                                                                            |
+----------v----------+                                                                                 |
|                     |                       -------------------------------------                     |
|     mountprobe     |     EventMessage                                                                |
|                     +---------------------->  udevevent.UdevEventMessageChannel  ---------------------+
|     listen loop     |
|                     |                       -------------------------------------
+---------------------+
```
