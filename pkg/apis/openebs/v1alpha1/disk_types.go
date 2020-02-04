/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// Disk is the Schema for the disks API
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec,omitempty"`
	Status DiskStatus `json:"status,omitempty"`
	Stats  DiskStat   `json:"stats,omitempty"`
	Device DeviceInfo `json:"deviceInfo"`
}

// DiskSpec defines the desired state of Disk
type DiskSpec struct {
	// Path contain devpath (e.g. /dev/sdb)
	Path string `json:"path"`

	// Capacity of the
	Capacity DiskCapacity `json:"capacity"`

	// Details contains static attributes (model, serial ..)
	Details DiskDetails `json:"details"`

	// DevLinks contains soft links of one disk
	DevLinks []DiskDevLink `json:"devlinks"`

	// Contains the data about filesystem on the disk
	FileSystem FileSystemInfo `json:"fileSystem,omitempty"`

	// Details of partitions in the disk (filesystem, partition type)
	PartitionDetails []Partition `json:"partitionDetails,omitempty"`
}

// DiskCapacity defines the physical and logical size of the disk
type DiskCapacity struct {
	// Storage is the disk capacity in bytes
	Storage uint64 `json:"storage"`

	// PhysicalSectorSize is disk physical-Sector size in bytes
	PhysicalSectorSize uint32 `json:"physicalSectorSize"`

	// LogicalSectorSize is disk logical-sector size in bytes
	LogicalSectorSize uint32 `json:"logicalSectorSize"`
}

// DiskDetails represent certain hardware/static attributes of the disk
type DiskDetails struct {
	// RotationRate is the rotation rate in RPM of the disk if not an SSD
	RotationRate uint16 `json:"rotationRate"`

	// DriveType represents the type of the drive like SSD, HDD etc.
	DriveType string `json:"driveType"`

	// Model is model of disk
	Model string `json:"model"`

	// Compliance is standards/specifications version implemented by device firmware
	//  such as SPC-1, SPC-2, etc
	Compliance string `json:"compliance"`

	// Serial is serial number of disk
	Serial string `json:"serial"`

	// Vendor is vendor of disk
	Vendor string `json:"vendor"`

	// FirmwareRevision is the disk firmware revision
	FirmwareRevision string `json:"firmwareRevision"`
}

// DiskDevLink holds the mapping between type and links like by-id type or by-path type link
type DiskDevLink struct {
	// Kind is the type of link like by-id or by-path.
	Kind string `json:"kind"`

	// Links are the soft links
	Links []string `json:"links"`
}

// Partition represents the partition information of the disk
type Partition struct {
	// PartitionType is the partition type of this partition.
	// LinuxLVM, SWAP, EFI are all partition types. They will be represented by
	// their corresponding code depending on the Partition table.
	//
	// Depending on the partition table on parent disk:
	// 1. For DOS partition table, two hexadecimal digits will be
	//    used for partition type
	// 2. For GPT partition table, a GUID will be present which corresponds to
	//    partition type in the format of `00000000-0000-0000-0000-000000000000`
	PartitionType string `json:"partitionType"`

	// FileSystem contains mountpoint and filesystem type
	FileSystem FileSystemInfo `json:"fileSystem,omitempty"`
}

// DiskStatus defines the observed state of Disk
type DiskStatus struct {
	// State is the current state of the disk (Active/Inactive)
	State DiskState `json:"state"`
}

// DiskState defines the observed state of the disk
type DiskState string

const (
	// DiskActive is the state for a physical disk that is connected to the node
	DiskActive DiskState = "Active"

	// DiskInactive is the state for a physical disk that is disconnected from a node
	DiskInactive DiskState = "Inactive"

	// DiskUnknown is the state for a physical disk whose state (attached/detached) cannot
	// be determined at this time.
	DiskUnknown DiskState = "Unknown"
)

// Temperature is the various temperature info reported by seachest
// about a physical disk. All info are in degree celsius
type Temperature struct {
	// CurrentTemperature is the current reported temperature of the drive
	CurrentTemperature int16 `json:"currentTemperature"`

	// HighestTemperature is the highest measured temperature of the drive in its lifetime
	HighestTemperature int16 `json:"highestTemperature"`

	// LowestTemperature is the lowest measured temperature of the drive in its lifetime
	LowestTemperature int16 `json:"lowestTemperature"`
}

// DiskStat gives variable attributes about the disk
type DiskStat struct {
	// TempInfo is the reported temperate of the disk
	TempInfo Temperature `json:"diskTemperature"`

	// TotalBytesRead
	// TODO @akhilerm document on what TotalBytesRead represent in a drive
	TotalBytesRead uint64 `json:"totalBytesRead"`

	// TotalBytesWritten
	// TODO @akhilerm document on what TotalBytesWritten represent in a drive
	TotalBytesWritten uint64 `json:"totalBytesWritten"`

	// DeviceUtilizationRate is utilization rate of the drive
	// ACS4 or SBC4 required for this to be valid
	DeviceUtilizationRate float64 `json:"deviceUtilizationRate"`

	// PercentEnduranceUsed
	// TODO @akhilerm document on what PercentEnduranceUsed represent in a drive
	PercentEnduranceUsed float64 `json:"percentEnduranceUsed"`
}

// DeviceInfo contains the info of the block device that is backed by this disk
type DeviceInfo struct {
	// DeviceUID is the NDM generated UID of the backed block device
	DeviceUID string `json:"blockDeviceUID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DiskList contains a list of Disk
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Disk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Disk{}, &DiskList{})
}
