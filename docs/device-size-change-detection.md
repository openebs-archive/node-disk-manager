# Blockdevice size change detection

## Introduction

This document presents the design for detecting changes in the capacity (or size) of
blocdevices in NDM.

## Summary

When a new blockdevice is detected, NDM collects all its details using different probes. One of the parameter collected is the size of the blockdevice. This is done using the **sysfs probe**. While the size of the device is detected when its first discovered, NDM remains oblivious to any changes that might happen to the size of the device once it has been scanned and added. Chaging the size of the device is quite common and easy to perfrom in cloud environments. Having out-dated info about the device size causes provisioning issues.Currently, the only way to update the blockdevice with the new size is to restart NDM.

This documents presents a system to detect changes to device capacity and update the upstream BlockDevice CR in etcd accordingly.

### Goals

- Detect changes to blockdevice capacity
- Retrieve new size and update the BlockDevice CR in etcd

## Design

The capacity of a blockdevice is currently fetched by the *sysfs* probe by reading the `size` attribute
of the blockdevice which is available as a file. This is only done when NDM first scans for devices or when it detects a new device has been added. NDM uses the *udev* system provided by the kernel events to capture events of devices being added or removed and act accordingly. The kernel uses the *udev* system to  inform the user-space programs about various events. These events include addition or removal of a blockdevice as well as events about any changes to a blockdevice. Whenever the size of a blockdevice changes, a change event is emitted by the kernel. Thus, size change detection is done by capturing change events from the *udev* system. However, change events are emitted for multiple reasons and not only a change in the device capacity. It has also been found that a significant amount of change events are emitted on VMs on cloud. This means that there's a large amount of noise that needs to be filtered out. While it is possible to filter out change events based on the subsystem (to recieve events from the blockdevice subsystem only), it is not possible to filter them futher to receive events only for blockdevice capacity change.

To capture change events from *udev*, the *sysfs* probe subscribes to change events using the existing *udevevent* package. On receiving a udev event, the sysfs probe processes the event to get the list of devices that were modified and creates a corresponding `controller.EventMessage` with the `requestedProbes` field containing only *sysfs* probe since only the size needs to be updated. This message is then pushed into `controller.EventMessageChannel` and further processed by the event handler.

## Alternative Design

A major issue with depending on *udev* for detecting size changes is that a good number of events won't be related to changes in the size. An alternative approach to detect changes to the size is to periodically read the `size` attribute file in *sysfs* and check if it has changed (polling). The polling interval can be fixed or configurable by user. A disadvantage of this approach is that polling will need to be done for the size file of each and every blockdevice that has been detected and depending on the poll interval, this operation can be expensive.

**Which approach to implement?**

Benchmarking the two approaches will provide a way to compare the two approaches. Implementing both the designs and giving the users an option to use the mechanism that suits their system best can also be done.
