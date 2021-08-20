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
	"errors"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

func (pe *ProbeEvent) changeBlockDevice(bd *blockdevice.BlockDevice, requestedProbes ...string) error {
	bdCopy := *bd
	haveEqualMountPoints := true
	pe.Controller.FillBlockDeviceDetails(bd, requestedProbes...)
	if bd.UUID == "" {
		uuid, ok := generateUUID(*bd)
		if !ok {
			klog.Error("could no generate uuid for device. aborting")
			return errors.New("could not indentify device uniquely")
		}
		bd.UUID = uuid
	}

	// Check if the mount-points have changed
	if len(bdCopy.FSInfo.MountPoint) != len(bd.FSInfo.MountPoint) {
		haveEqualMountPoints = false
	} else {
		for _, mountPoint := range bd.FSInfo.MountPoint {
			if !util.Contains(bdCopy.FSInfo.MountPoint, mountPoint) {
				haveEqualMountPoints = false
				break
			}
		}
	}
	/*
	 * Change detection is only employed for detecting changes in:
	 * 1. Size
	 * 2. Filesystem
	 * 3. Mount-points
	 *
	 * Check if any of these have actually changed. This prevents unnecessary
	 * calls to the k8s api server.
	 */
	if bdCopy.Capacity.Storage == bd.Capacity.Storage &&
		bdCopy.FSInfo.FileSystem == bd.FSInfo.FileSystem &&
		haveEqualMountPoints {
		klog.Infof("no changes in %s. Skipping update", bd.DevPath)
		return nil
	}

	pe.addBlockDeviceToHierarchyCache(*bd)
	if !pe.Controller.ApplyFilter(bd) {
		return nil
	}
	apiBlockdevice := pe.Controller.NewDeviceInfoFromBlockDevice(bd).ToDevice()
	apiBlockdevice.SetNamespace(pe.Controller.Namespace)
	klog.Info("updating bd: ", apiBlockdevice.GetName())
	return pe.Controller.UpdateBlockDevice(apiBlockdevice, nil)
}
