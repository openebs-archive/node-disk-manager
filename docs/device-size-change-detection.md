# Blockdevice size change detection

## Introduction

This document presents the design for detecting changes in the capacity (or size) of
blockdevices in NDM.

## Summary

When a new blockdevice is detected, NDM collects all its details using different probes. One of the parameter collected is the size of the blockdevice. This is done using the **sysfs probe**. While the size of the device is detected when its first discovered, NDM remains oblivious to any changes that might happen to the size of the device once it has been scanned and added. Chaging the size of the device is quite common and easy to perfrom in cloud environments. Having out-dated info about the device size causes provisioning issues.Currently, the only way to update the blockdevice with the new size is to restart NDM.

This documents presents a system to detect changes to device capacity and update the upstream BlockDevice CR in etcd accordingly.

### Goals

- Detect changes to blockdevice capacity
- Retrieve new size and update the BlockDevice CR in etcd

## Design

The capacity of a blockdevice is currently fetched by the *sysfs-probe* by reading the `size` attribute
of the blockdevice which is available as a file. This is only done when NDM first scans for devices or when it detects a new device has been added. NDM uses the *udev* system provided by the kernel to capture events related to blockdevices. The kernel uses the *udev* system to inform the user-space programs about various events. These events include addition or removal of a blockdevice as well as events about any changes to a blockdevice. Whenever the size of a blockdevice changes, a change event is emitted by the kernel. Thus, size change detection is done by capturing change events from the *udev* system. However, change events are emitted for multiple reasons and not only a change in the device capacity. It has also been found that a significant amount of change events are emitted on VMs on cloud. This means that there's a large amount of noise that needs to be filtered out. While it is possible to filter out change events based on the subsystem (to recieve events from the blockdevice subsystem only), it is not possible to filter them futher to receive events only for blockdevice capacity change.

The udev change events are captured and processed by *udev-probe*. On recieveing a change event, *udev-probe* extracts the dev path for the changed device from the udev event and pushes a new [`EventMessage`](https://github.com/openebs/node-disk-manager/blob/eb1aa4b63a9fe0b3e6ef555e1c9528f095cb8680/cmd/ndm_daemonset/controller/probe.go#L29) to the existing event message channel ([`controller.EventMessageChannel`](https://github.com/openebs/node-disk-manager/blob/eb1aa4b63a9fe0b3e6ef555e1c9528f095cb8680/cmd/ndm_daemonset/controller/probe.go#L36)). The probes run by the controller for this message are limited to *udev-probe* (for detecting changes to the device filesystem which may not be detected by the [mount-change detection system](https://github.com/openebs/node-disk-manager/blob/eb1aa4b63a9fe0b3e6ef555e1c9528f095cb8680/docs/mount-change-detection.md)) and *sysfs-probe* (for getting changes to the device size) by setting the `RequestedProbes` field of the event message. The action of the event message is set to `ChangeEA` (change event action) to direct the event to change handler.

### Change handler updates

To reduce the amount of update requests to the Kubernetes API server, a check is added to the change handler to send a request only if the following properties have changed in the blockdevice:

- size
- mount points
- filesystem

The existing copy of the blockdevice in controller cache is compared with the copy of blockdevice that is run through the requested probe for determining if any of the above properties actually changed. This check is important since there can be a lot of change events in udev unrelated to changes in the above properties which eventually end up in change handler.

## Alternate Design

A major issue with depending on *udev* for detecting size changes is that a good number of events won't be related to changes in the size. An alternative approach to detect changes to the size is to periodically read the `size` attribute file in *sysfs* and check if it has changed (polling). A disadvantage of this approach is that polling will need to be done for the size file of each and every blockdevice that has been detected and depending on the poll interval, this operation can be expensive.

However, from performance tests and profiling it was found that both the approaches (udev and sysfs polling) have very similar performance and don't take up a lot of resources. A major proportion of NDM's CPU time is comsumed in communicating with the Kubernetes API server. Since NDM already uses the udev system for device additions and deletions, enabling change detection with udev is the better choice as the existing system can be used. Also, the change events are generated for changes to several other properties of a blockdevice. This allows for future extension if change detection for more properties is required.
