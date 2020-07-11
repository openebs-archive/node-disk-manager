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
	"github.com/openebs/node-disk-manager/blockdevice"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"

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
func (pe *ProbeEvent) deleteBlockDevice(bd blockdevice.BlockDevice, bdAPIList *apis.BlockDeviceList) error {

	if !pe.removeBlockDeviceFromHierarchyCache(bd) {
		return nil
	}

	uuid, ok := generateUUID(bd)
	// this can happen if the device didn't have any unique identifiers
	if !ok {
		klog.Info("could not recreate UUID while removing device")
		return nil
	}

	existingBlockDeviceResource := pe.Controller.GetExistingBlockDeviceResource(bdAPIList, uuid)
	if existingBlockDeviceResource == nil {
		return nil
	}
	pe.Controller.DeactivateBlockDevice(*existingBlockDeviceResource)

	return nil
}
