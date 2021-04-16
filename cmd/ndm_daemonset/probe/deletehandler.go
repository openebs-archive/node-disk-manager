/*
Copyright 2020 The OpenEBS Authors

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
	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"

	"k8s.io/klog"
)

// removeBlockDeviceFromHierarchyCache removes a block device from the hierarchy.
// returns true if the device existed in the cache, else returns false
func (pe *ProbeEvent) removeBlockDeviceFromHierarchyCache(bd blockdevice.BlockDevice) bool {
	_, ok := pe.Controller.BDHierarchy[bd.DevPath]
	if !ok {
		klog.Infof("Disk %s not in hierarchy", bd.DevPath)
		// not in hierarchy continue
		return false
	}
	// remove from the hierarchy
	delete(pe.Controller.BDHierarchy, bd.DevPath)
	return true
}

// deleteBlockDevice marks the block device resource as inactive
// The following cases are handled
//	1. Device using legacy UUID
//	2. Device using GPT UUID
//	3. Device using partition table UUID (zfs localPV)
//  4. Device using the partition table / fs uuid annotation
func (pe *ProbeEvent) deleteBlockDevice(bd blockdevice.BlockDevice, bdAPIList *apis.BlockDeviceList) error {

	if !pe.removeBlockDeviceFromHierarchyCache(bd) {
		return nil
	}

	// try with gpt uuid
	if uuid, ok := generateUUID(bd); ok {
		existingBD := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, uuid)
		if existingBD != nil {
			pe.Controller.DeactivateBlockDevice(*existingBD)
			klog.V(4).Infof("deactivated device: %s, using GPT UUID", bd.DevPath)
			return nil
		}
		// uuid could be generated, but the disk may be using the legacy scheme
	}

	// try with partition table uuid - for zfs local pV
	if partUUID, ok := generateUUIDFromPartitionTable(bd); ok {
		existingBD := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, partUUID)
		if existingBD != nil {
			pe.Controller.DeactivateBlockDevice(*existingBD)
			klog.V(4).Infof("deactivated device: %s, using partition table UUID", bd.DevPath)
			return nil
		}
	}

	// try with FSUUID annotation
	if existingBD := getExistingBDWithFsUuid(bd, bdAPIList); existingBD != nil {
		pe.Controller.DeactivateBlockDevice(*existingBD)
		klog.V(4).Infof("deactivated device: %s, using FS UUID annotation", bd.DevPath)
		return nil
	}

	// try with partition uuid annotation
	// if the device is a partition, the parent can get deactivated because of search by partition UUID.
	// Therefore the search result is used only if the device is not a partition.
	if existingBD := getExistingBDWithPartitionUUID(bd, bdAPIList); bd.DeviceAttributes.DeviceType != blockdevice.BlockDeviceTypePartition &&
		existingBD != nil {
		pe.Controller.DeactivateBlockDevice(*existingBD)
		klog.V(4).Infof("deactivated device: %s, using Partition UUID annotation", bd.DevPath)
		return nil
	}

	// try with legacy uuid
	legacyUUID, _ := generateLegacyUUID(bd)
	existingBD := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, legacyUUID)
	if existingBD != nil {
		pe.Controller.DeactivateBlockDevice(*existingBD)
		klog.V(4).Infof("deactivated device: %s, using legacy UUID", bd.DevPath)
		return nil
	}

	return nil
}
