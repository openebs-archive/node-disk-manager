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
	// bdAPIList is the list of all the BlockDevice resources in the cluster
	bdAPIList, err := pe.Controller.ListBlockDeviceResource(true)
	if err != nil {
		klog.Error(err)
		go Rescan(pe.Controller)
		return
	}

	isGPTBasedUUIDEnabled := features.FeatureGates.IsEnabled(features.GPTBasedUUID)

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

		if isGPTBasedUUIDEnabled {
			err := pe.addBlockDevice(*device, bdAPIList)
			if err != nil {
				isErrorDuringUpdate = true
				klog.Error(err)
				// if error occurs we should start the scan again
				break
			}
		} else {
			// if GPTBasedUUID is disabled and the device type is partition,
			// the event can be skipped.
			if device.DeviceAttributes.DeviceType == libudevwrapper.UDEV_PARTITION {
				klog.Info("GPTBasedUUID disabled. skip creating block device resource for partition.")
				continue
			}
			deviceInfo := pe.Controller.NewDeviceInfoFromBlockDevice(device)

			existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, deviceInfo.UUID)
			err := pe.Controller.PushBlockDeviceResource(existingBlockDeviceResource, deviceInfo)
			if err != nil {
				isErrorDuringUpdate = true
				klog.Error(err)
			}
		}
	}

	if isErrorDuringUpdate {
		go Rescan(pe.Controller)
	}
}

// deleteBlockDeviceEvent deactivate blockdevice resource using uuid from etcd
func (pe *ProbeEvent) deleteBlockDeviceEvent(msg controller.EventMessage) {
	bdAPIList, err := pe.Controller.ListBlockDeviceResource(false)
	if err != nil {
		klog.Error(err)
	}

	isDeactivated := true
	isGPTBasedUUIDEnabled := features.FeatureGates.IsEnabled(features.GPTBasedUUID)

	for _, device := range msg.Devices {
		if isGPTBasedUUIDEnabled {
			_ = pe.deleteBlockDevice(*device, bdAPIList)
		} else {
			existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, device.UUID)
			if existingBlockDeviceResource == nil {
				// do nothing, may be the disk was filtered, or it was not created
				isDeactivated = false
				continue
			}
			pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)
		}

	}

	// rescan only if GPT based UUID is disabled.
	if !isDeactivated && !isGPTBasedUUIDEnabled {
		go Rescan(pe.Controller)
	}
}

// generateUUID creates a new UUID based on the algorithm proposed in
// https://github.com/openebs/openebs/pull/2666
func generateUUID(bd blockdevice.BlockDevice) (string, bool) {
	var ok bool
	var uuidField, uuid string

	// select the field which is to be used for generating UUID
	//
	// Serial number is not used directly for UUID generation. This is because serial number is not
	// unique in some cloud environments. For example, in GCP the serial number is
	// configurable by the --device-name flag while attaching the disk.
	// If this flag is not provided, GCP automatically assigns the serial number
	// which is unique only to the node. Therefore Serial number is used only in cases
	// where the disk has a WWN.
	//
	// If disk has WWN, a combination of WWN+Serial will be used. This is done because there are cases
	// where the disks has same WWN but different serial. It is seen in some storage arrays.
	// All the LUNs will have same WWN, but different serial.
	//
	// PartitionTableUUID is not used for UUID generation in NDM. The only case where the disk has a PartitionTable
	// and not partition is when, the user has manually created a partition table without writing any actual partitions.
	// This means NDM will have to give its consumers the entire disk, i.e consumers will have access to the sectors
	// where partition table is written. If consumers decide to reformat or erase the disk completely the partition
	// table UUID is also lost, making NDM unable to identify the disk. Hence, even if a partition table is present
	// NDM will rewrite it and create a new GPT table and a single partition. Thus consumers will have access only to
	// the partition and the unique data will be stored in sectors where consumers do not have access.

	switch {
	case bd.DeviceAttributes.DeviceType == blockdevice.BlockDeviceTypePartition:
		// The partition entry UUID is used when a partition (/dev/sda1) is processed. The partition UUID should be used
		// if available, other than the partition table UUID, because multiple partitions can have the same partition table
		// UUID, but each partition will have a different UUID.
		klog.Infof("device(%s) is a partition, using partition UUID: %s", bd.DevPath, bd.PartitionInfo.PartitionEntryUUID)
		uuidField = bd.PartitionInfo.PartitionEntryUUID
		ok = true
	case len(bd.DeviceAttributes.WWN) > 0:
		// if device has WWN, both WWN and Serial will be used for UUID generation.
		klog.Infof("device(%s) has a WWN, using WWN: %s and Serial: %s",
			bd.DevPath,
			bd.DeviceAttributes.WWN, bd.DeviceAttributes.Serial)
		uuidField = bd.DeviceAttributes.WWN +
			bd.DeviceAttributes.Serial
		ok = true
	case len(bd.FSInfo.FileSystemUUID) > 0:
		klog.Infof("device(%s) has a filesystem, using filesystem UUID: %s", bd.DevPath, bd.FSInfo.FileSystemUUID)
		uuidField = bd.FSInfo.FileSystemUUID
		ok = true
	}

	if ok {
		uuid = blockdevice.BlockDevicePrefix + util.Hash(uuidField)
		klog.Infof("generated uuid: %s for device: %s", uuid, bd.DevPath)
	}

	return uuid, ok
}
