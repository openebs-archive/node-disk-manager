/*
Copyright 2019 The OpenEBS Authors.

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

package controller

import (
	bd "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/udev"
)

// NewDeviceInfoFromBlockDevice converts the internal BlockDevice struct to
// the BlockDevice API resource
func (c *Controller) NewDeviceInfoFromBlockDevice(blockDevice *bd.BlockDevice) *DeviceInfo {

	deviceDetails := NewDeviceInfo()

	deviceDetails.NodeAttributes = make(map[string]string)

	// copying node attributes from core blockdevice to device info
	for k, v := range blockDevice.NodeAttributes {
		deviceDetails.NodeAttributes[k] = v
	}

	deviceDetails.UUID = blockDevice.UUID
	deviceDetails.Labels = blockDevice.Labels
	deviceDetails.Capacity = blockDevice.Capacity.Storage
	deviceDetails.Model = blockDevice.DeviceAttributes.Model
	deviceDetails.Serial = blockDevice.DeviceAttributes.Serial
	deviceDetails.Vendor = blockDevice.DeviceAttributes.Vendor
	deviceDetails.Path = blockDevice.DevPath
	deviceDetails.FirmwareRevision = blockDevice.DeviceAttributes.FirmwareRevision

	for _, devlink := range blockDevice.DevLinks {
		if devlink.Kind == udev.BY_ID_LINK {
			deviceDetails.ByIdDevLinks = devlink.Links
		} else if devlink.Kind == udev.BY_PATH_LINK {
			deviceDetails.ByPathDevLinks = devlink.Links
		}
	}
	deviceDetails.LogicalBlockSize = blockDevice.DeviceAttributes.LogicalBlockSize
	deviceDetails.PhysicalBlockSize = blockDevice.DeviceAttributes.PhysicalBlockSize
	deviceDetails.HardwareSectorSize = blockDevice.DeviceAttributes.HardwareSectorSize
	deviceDetails.DriveType = blockDevice.DeviceAttributes.DriveType
	deviceDetails.DeviceType = blockDevice.DeviceAttributes.DeviceType

	deviceDetails.Compliance = blockDevice.DeviceAttributes.Compliance
	deviceDetails.FileSystemInfo.FileSystem = blockDevice.FSInfo.FileSystem
	// currently only the first mount point will be taken.
	if len(blockDevice.FSInfo.MountPoint) != 0 {
		deviceDetails.FileSystemInfo.MountPoint = blockDevice.FSInfo.MountPoint[0]
	}
	return deviceDetails
}
