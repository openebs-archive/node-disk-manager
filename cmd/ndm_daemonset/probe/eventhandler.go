/*
Copyright 2018 OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/features"
	"github.com/openebs/node-disk-manager/pkg/partition"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

// EventAction action type for disk events like attach or detach events
type EventAction string

const (
	// AttachEA is attach disk event name
	AttachEA EventAction = libudevwrapper.UDEV_ACTION_ADD
	// DetachEA is detach disk event name
	DetachEA EventAction = libudevwrapper.UDEV_ACTION_REMOVE
)

// ProbeEvent struct contain a copy of controller it will update disk resources
type ProbeEvent struct {
	Controller *controller.Controller
}

// addBlockDeviceEvent fill block device details from different probes and push it to etcd
func (pe *ProbeEvent) addBlockDeviceEvent(msg controller.EventMessage) {
	bdList, err := pe.Controller.ListBlockDeviceResource()
	if err != nil {
		klog.Error(err)
		go pe.initOrErrorEvent()
		return
	}
	isErrorDuringUpdate := false
	// iterate through each block device and perform the add/update operation
	for _, device := range msg.Devices {
		klog.Infof("Processing details for %s", device.DevPath)
		pe.Controller.FillBlockDeviceDetails(device)
		// if ApplyFilter returns true then we process the event further
		if !pe.Controller.ApplyFilter(device) {
			continue
		}
		klog.Infof("Processed details for %s", device.DevPath)

		// if GPTBasedUUID need to be used, generate the UUID,
		// if UUID cannot be generated create a GPT partition
		if pe.Controller.FeatureGates.IsEnabled(features.GPTBasedUUID) {
			if len(device.Partitions) > 0 {
				klog.Info("device has partitions. not creating blockdevice resource")
				continue
			}
			if len(device.Holders) > 0 {
				klog.Info("device has holder devices, not creating blockdevice resource")
				continue
			}
			uuid, ok := generateUUID(*device)
			// manaully create a single partition on the device
			if !ok {
				klog.Info("starting to create partitions")
				d := partition.Disk{
					DevPath:          device.DevPath,
					DiskSize:         device.Capacity.Storage,
					LogicalBlockSize: uint64(device.DeviceAttributes.LogicalBlockSize),
				}
				if err := d.CreatePartitionTable(); err != nil {
					klog.Errorf("error creating partition table for %s, %v", device.DevPath, err)
					continue
				}
				if err = d.AddPartition(); err != nil {
					klog.Errorf("error creating partition for %s, %v", device.DevPath, err)
					continue
				}
				if err = d.ApplyPartitionTable(); err != nil {
					klog.Errorf("error writing partition data to %s, %v", device.DevPath, err)
					continue
				}
				klog.Infof("created new partition in %s", device.DevPath)
				continue
			}
			klog.Infof("generated UUID: %s for device: %s", uuid, device.DevPath)
			device.UUID = uuid
		}

		deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(device)

		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdList, deviceInfo.UUID)
		err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
		if err != nil {
			isErrorDuringUpdate = true
			klog.Error(err)
		}
	}

	if isErrorDuringUpdate {
		go pe.initOrErrorEvent()
	}
}

// deleteBlockDeviceEvent deactivate blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteBlockDeviceEvent(msg controller.EventMessage) {
	bdList, err := pe.Controller.ListBlockDeviceResource()
	if err != nil {
		klog.Error(err)
	}

	isDeactivated := true
	isGPTBasedUUIDEnabled := pe.Controller.FeatureGates.IsEnabled(features.GPTBasedUUID)

	for _, device := range msg.Devices {
		// create the new UUID for removing the device
		if isGPTBasedUUIDEnabled {
			uuid, ok := generateUUID(*device)
			if !ok {
				klog.Info("could not recreate UUID while removing device")
				continue
			}
			// use the new UUID for deactivating the blockdevice
			device.UUID = uuid
		}

		existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdList, device.UUID)
		if existingBlockDeviceResource == nil {
			isDeactivated = false
			continue
		}
		pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)
	}

	// rescan only if GPT based UUID is disabled.
	if !isDeactivated && !isGPTBasedUUIDEnabled {
		go pe.initOrErrorEvent()
	}
}

// initOrErrorEvent rescan system and update disk resource this is
// used for initial setup and when any uid mismatch or error occurred.
func (pe *ProbeEvent) initOrErrorEvent() {
	udevProbe := newUdevProbe(pe.Controller)
	defer udevProbe.free()
	err := udevProbe.scan()
	if err != nil {
		klog.Error(err)
	}
}

// generateUUID creates a new UUID based on the algorithm proposed in
// https://github.com/openebs/openebs/pull/2666
func generateUUID(bd blockdevice.BlockDevice) (string, bool) {
	var ok bool
	var uuidField, uuid string

	// select the field which is to be used for generating UUID
	switch {
	case bd.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION:
		klog.Infof("device(%s) is a partition, using partition UUID: %s", bd.DevPath, bd.PartitionInfo.PartitionEntryUUID)
		uuidField = bd.PartitionInfo.PartitionEntryUUID
		ok = true
	case len(bd.DeviceAttributes.WWN) > 0:
		klog.Infof("device(%s) has a WWN, using WWN: %s", bd.DevPath, bd.DeviceAttributes.WWN)
		uuidField = bd.DeviceAttributes.WWN
		ok = true
	case len(bd.PartitionInfo.PartitionTableUUID) > 0:
		klog.Infof("device(%s) has a partition table, using partition table UUID: %s", bd.DevPath, bd.PartitionInfo.PartitionTableUUID)
		uuidField = bd.PartitionInfo.PartitionTableUUID
		ok = true
	case len(bd.FSInfo.FileSystemUUID) > 0:
		klog.Infof("device(%s) has a filesystem, using filesystem UUID: %s", bd.DevPath, bd.FSInfo.FileSystemUUID)
		uuidField = bd.FSInfo.FileSystemUUID
		ok = true
	}

	if ok {
		uuid = libudevwrapper.NDMBlockDevicePrefix + util.Hash(uuidField)
		klog.Infof("generated uuid: %s for device: %s", uuid, bd.DevPath)
	}

	return uuid, ok
}
