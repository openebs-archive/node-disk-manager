/*
Copyright 2019 OpenEBS Authors.

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
	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	bd "github.com/openebs/node-disk-manager/blockdevice"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceInfo contains details of blockdevice which can be converted into api.BlockDevice
// There will be one DeviceInfo created for each physical disk (may change in
// future). At the end it is converted to BlockDevice struct which will be pushed to
// etcd as a CR of that blockdevice.
type DeviceInfo struct {
	// NodeAttributes is the attributes of the node to which this block device is attached,
	// like hostname, nodename
	NodeAttributes bd.NodeAttribute
	// Optional labels that can be added to the blockdevice resource
	Labels             map[string]string
	UUID               string   // UUID of backing disk
	Capacity           uint64   // Capacity of blockdevice
	Model              string   // Do blockdevice have model ??
	Serial             string   // Do blockdevice have serial no ??
	Vendor             string   // Vendor of blockdevice
	Path               string   // blockdevice Path like /dev/sda
	ByIdDevLinks       []string // ByIdDevLinks contains by-id devlinks
	ByPathDevLinks     []string // ByPathDevLinks contains by-path devlinks
	FirmwareRevision   string   // FirmwareRevision is the firmware revision for a disk
	LogicalBlockSize   uint32   // LogicalBlockSize is the logical block size of the device in bytes
	PhysicalBlockSize  uint32   // PhysicalBlockSize is the physical block size in bytes
	HardwareSectorSize uint32   // HardwareSectorSize is the hardware sector size in bytes
	Compliance         string   // Compliance is implemented specifications version i.e. SPC-1, SPC-2, etc
	DeviceType         string   // DeviceType represents the type of device, like disk/sparse/partition
	DriveType          string   // DriveType represents the type of backing drive HDD/SSD
	PartitionType      string   // Partition type if the blockdevice is a partition
	FileSystemInfo     FSInfo   // FileSystem info of the blockdevice like FSType and MountPoint
}

// NewDeviceInfo returns a pointer of empty DeviceInfo
// struct which will be field from DiskInfo.
func NewDeviceInfo() *DeviceInfo {
	deviceInfo := &DeviceInfo{}
	return deviceInfo
}

// FSInfo defines the filesystem related information of block device/disk, like mountpoint and
// filesystem
type FSInfo struct {
	FileSystem string // Filesystem on the block device
	MountPoint string // MountPoint of the block device
}

// ToDevice convert deviceInfo struct to api.BlockDevice
// type which will be pushed to etcd
func (di *DeviceInfo) ToDevice() apis.BlockDevice {
	blockDevice := apis.BlockDevice{}
	blockDevice.Spec = di.getDeviceSpec()
	blockDevice.ObjectMeta = di.getObjectMeta()
	blockDevice.TypeMeta = di.getTypeMeta()
	blockDevice.Status = di.getStatus()
	return blockDevice
}

// getObjectMeta returns ObjectMeta struct which contains
// labels and Name of resource. It is used to populate data
// of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getObjectMeta() metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Name:        di.UUID,
	}
	objectMeta.Labels[KubernetesHostNameLabel] = di.NodeAttributes[HostNameKey]
	objectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	objectMeta.Labels[NDMManagedKey] = TrueString
	// adding custom labels
	for k, v := range di.Labels {
		objectMeta.Labels[k] = v
	}
	return objectMeta
}

// getTypeMeta returns TypeMeta struct which contains
// resource kind and version. It is used to populate
// data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getTypeMeta() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       NDMBlockDeviceKind,
		APIVersion: NDMVersion,
	}
	return typeMeta
}

// getStatus returns DeviceStatus struct which contains
// state of BlockDevice resource. It is used to populate data
// of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getStatus() apis.DeviceStatus {
	deviceStatus := apis.DeviceStatus{
		ClaimState: apis.BlockDeviceUnclaimed,
		State:      NDMActive,
	}
	return deviceStatus
}

// getDiskSpec returns DiskSpec struct which contains info of blockdevice like :
// - path - /dev/sdb etc.
// - capacity - (size,logical sector size ...)
// - devlinks - (by-id , by-path links)
// It is used to populate data of BlockDevice struct of blockdevice CR.
func (di *DeviceInfo) getDeviceSpec() apis.DeviceSpec {
	deviceSpec := apis.DeviceSpec{}
	deviceSpec.NodeAttributes.NodeName = di.NodeAttributes[NodeNameKey]
	deviceSpec.Path = di.getPath()
	deviceSpec.Details = di.getDeviceDetails()
	deviceSpec.Capacity = di.getDeviceCapacity()
	deviceSpec.DevLinks = di.getDeviceLinks()
	deviceSpec.Partitioned = NDMNotPartitioned
	deviceSpec.FileSystem = di.FileSystemInfo.getFileSystemInfo()
	return deviceSpec
}

// getPath returns path of the blockdevice like (/dev/sda , /dev/sdb ...).
// It is used to populate data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getPath() string {
	return di.Path
}

// getDeviceDetails returns DeviceDetails struct which contains primary
// and static info of blockdevice resource like model, serial, vendor etc.
// It is used to populate data of BlockDevice struct which of BlockDevice CR.
func (di *DeviceInfo) getDeviceDetails() apis.DeviceDetails {
	deviceDetails := apis.DeviceDetails{}
	deviceDetails.Model = di.Model
	deviceDetails.Serial = di.Serial
	deviceDetails.Vendor = di.Vendor
	deviceDetails.FirmwareRevision = di.FirmwareRevision
	deviceDetails.Compliance = di.Compliance
	deviceDetails.DeviceType = di.DeviceType
	deviceDetails.DriveType = di.DriveType
	deviceDetails.LogicalBlockSize = di.LogicalBlockSize
	deviceDetails.PhysicalBlockSize = di.PhysicalBlockSize
	deviceDetails.HardwareSectorSize = di.HardwareSectorSize

	return deviceDetails
}

// getDiskCapacity returns DeviceCapacity struct which contains:
// -size of disk (in bytes)
// -logical sector size (in bytes)
// -physical sector size (in bytes)
// It is used to populate data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getDeviceCapacity() apis.DeviceCapacity {
	capacity := apis.DeviceCapacity{}
	capacity.Storage = di.Capacity
	capacity.LogicalSectorSize = di.LogicalBlockSize
	capacity.PhysicalSectorSize = di.PhysicalBlockSize
	return capacity
}

// getDiskLinks returns DeviceDevLink struct which contains
// soft links like by-id ,by-path link. It is used to populate
// data of BlockDevice struct of BlockDevice CR.
func (di *DeviceInfo) getDeviceLinks() []apis.DeviceDevLink {
	devLinks := make([]apis.DeviceDevLink, 0)
	if len(di.ByIdDevLinks) != 0 {
		byIDLinks := apis.DeviceDevLink{
			Kind:  "by-id",
			Links: di.ByIdDevLinks,
		}
		devLinks = append(devLinks, byIDLinks)
	}
	if len(di.ByPathDevLinks) != 0 {
		byPathLinks := apis.DeviceDevLink{
			Kind:  "by-path",
			Links: di.ByPathDevLinks,
		}
		devLinks = append(devLinks, byPathLinks)
	}
	return devLinks
}

func (fs *FSInfo) getFileSystemInfo() apis.FileSystemInfo {
	fsInfo := apis.FileSystemInfo{}
	fsInfo.Type = fs.FileSystem
	fsInfo.Mountpoint = fs.MountPoint
	return fsInfo
}
