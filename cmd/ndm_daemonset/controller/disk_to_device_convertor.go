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
	"github.com/openebs/node-disk-manager/pkg/udev"
	"strings"
)

func (c *Controller) NewDeviceInfoFromDiskInfo(diskDetails *DiskInfo) *DeviceInfo {

	deviceDetails := NewDeviceInfo()

	deviceDetails.HostName = c.HostName
	Uuid := strings.TrimPrefix(diskDetails.ProbeIdentifiers.Uuid, udev.NDMDiskPrefix)
	deviceDetails.UUID = udev.NDMBlockDevicePrefix + Uuid
	deviceDetails.Capacity = diskDetails.Capacity
	deviceDetails.Model = diskDetails.Model
	deviceDetails.Serial = diskDetails.Serial
	deviceDetails.Vendor = diskDetails.Vendor
	deviceDetails.Path = diskDetails.Path
	deviceDetails.ByIdDevLinks = diskDetails.ByIdDevLinks
	deviceDetails.ByPathDevLinks = diskDetails.ByPathDevLinks
	deviceDetails.LogicalSectorSize = diskDetails.LogicalSectorSize
	deviceDetails.PhysicalSectorSize = diskDetails.PhysicalSectorSize
	deviceDetails.Compliance = diskDetails.Compliance
	deviceDetails.DeviceType = diskDetails.DriveType
	deviceDetails.FileSystemInfo.FileSystem = diskDetails.FileSystemInformation.FileSystem
	deviceDetails.FileSystemInfo.MountPoint = diskDetails.FileSystemInformation.MountPoint
	return deviceDetails
}
