/*
Copyright 2021 The OpenEBS Authors

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// +kubebuilder:resource:scope=Namespaced,shortName=bd

// BlockDevice is the Schema used to represent a BlockDevice CR
// +kubebuilder:printcolumn:name="NodeName",type="string",JSONPath=`.spec.nodeAttributes.nodeName`
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=`.spec.path`,priority=1
// +kubebuilder:printcolumn:name="FSType",type="string",JSONPath=`.spec.filesystem.fsType`,priority=1
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=`.spec.capacity.storage`
// +kubebuilder:printcolumn:name="ClaimState",type="string",JSONPath=`.status.claimState`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:resource:scope=Namespaced,shortName=bd
type BlockDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlockDeviceSpec   `json:"spec,omitempty"`
	Status BlockDeviceStatus `json:"status,omitempty"`
}

// BlockDeviceSpec defines the properties and runtime status of a BlockDevice
type BlockDeviceSpec struct {

	// AggregateDevice was intended to store the hierachical
	// information in cases of LVM. However this is currently
	// not implemented and may need to be re-looked into for
	// better design. To be deprecated
	// +optional
	AggregateDevice string `json:"aggregateDevice,omitempty"`

	// Capacity
	Capacity DeviceCapacity `json:"capacity"`

	// ClaimRef is the reference to the BDC which has claimed this BD
	// +optional
	ClaimRef *v1.ObjectReference `json:"claimRef,omitempty"`

	// Details contain static attributes of BD like model,serial, and so forth
	// +optional
	Details DeviceDetails `json:"details"`

	// DevLinks contains soft links of a block device like
	// /dev/by-id/...
	// /dev/by-uuid/...
	DevLinks []DeviceDevLink `json:"devlinks"`

	// FileSystem contains mountpoint and filesystem type
	// +optional
	FileSystem FileSystemInfo `json:"filesystem,omitempty"`

	// NodeAttributes has the details of the node on which BD is attached
	NodeAttributes NodeAttribute `json:"nodeAttributes"`

	// ParentDevice was intended to store the UUID of the parent
	// Block Device as is the case for partitioned block devices.
	//
	// For example:
	// /dev/sda is the parent for /dev/sda1
	// To be deprecated
	// +kubebuilder:validation:Pattern:`^/dev/[a-z]{3,4}$`
	// +optional
	ParentDevice string `json:"parentDevice,omitempty"`

	// Partitioned represents if BlockDevice has partitions or not (Yes/No)
	// Currently always default to No.
	// To be deprecated
	// +kubebuilder:validation:Enum:=Yes;No
	// +optional
	Partitioned string `json:"partitioned"`

	// Path contain devpath (e.g. /dev/sdb)
	// +kubebuilder:validation:Pattern:`^/dev/[a-z]{3,4}$`
	Path string `json:"path"`
}

// NodeAttribute defines the attributes of a node where
// the block device is attached.
//
// Note: Prior to introducing NodeAttributes, the BD would
// only support gathering hostname and add it as a label
// to the BD resource.
//
// In some use cases, the caller has access only to node name, not
// the hostname. node name and hostname are different in certain
// Kubernetes clusters.
//
// NodeAttributes is added to contain attributes that are not
// available on the labels like - node name, uuid, etc.
//
// The node attributes are helpful in querying for block devices
// based on node attributes.  Also, adding this in the spec allows for
// displaying in node name in the `kubectl get bd`
//
// Capture and add nodeUUID to BD, that can help in determining
// if the node was recreated with same node name.
type NodeAttribute struct {
	// NodeName is the name of the Kubernetes node resource on which the device is attached
	// +optional
	NodeName string `json:"nodeName"`
}

// DeviceCapacity defines the physical and logical size of the block device
type DeviceCapacity struct {
	// Storage is the blockdevice capacity in bytes
	Storage uint64 `json:"storage"`

	// PhysicalSectorSize is blockdevice physical-Sector size in bytes
	// +optional
	PhysicalSectorSize uint32 `json:"physicalSectorSize"`

	// LogicalSectorSize is blockdevice logical-sector size in bytes
	// +optional
	LogicalSectorSize uint32 `json:"logicalSectorSize"`
}

// DeviceDetails represent certain hardware/static attributes of the block device
type DeviceDetails struct {
	// DeviceType represents the type of device like
	// sparse, disk, partition, lvm, crypt
	// +kubebuilder:validation:Enum:=disk;partition;sparse;loop;lvm;crypt;dm;mpath
	// +optional
	DeviceType string `json:"deviceType"`

	// DriveType is the type of backing drive, HDD/SSD
	// +kubebuilder:validation:Enum:=HDD;SSD;Unknown;""
	// +optional
	DriveType string `json:"driveType"`

	// LogicalBlockSize is the logical block size in bytes
	// reported by /sys/class/block/sda/queue/logical_block_size
	// +optional
	LogicalBlockSize uint32 `json:"logicalBlockSize"`

	// PhysicalBlockSize is the physical block size in bytes
	// reported by /sys/class/block/sda/queue/physical_block_size
	// +optional
	PhysicalBlockSize uint32 `json:"physicalBlockSize"`

	// HardwareSectorSize is the hardware sector size in bytes
	// +optional
	HardwareSectorSize uint32 `json:"hardwareSectorSize"`

	// Model is model of disk
	// +optional
	Model string `json:"model"`

	// Compliance is standards/specifications version implemented by device firmware
	//  such as SPC-1, SPC-2, etc
	// +optional
	Compliance string `json:"compliance"`

	// Serial is serial number of disk
	// +optional
	Serial string `json:"serial"`

	// Vendor is vendor of disk
	// +optional
	Vendor string `json:"vendor"`

	// FirmwareRevision is the disk firmware revision
	// +optional
	FirmwareRevision string `json:"firmwareRevision"`
}

// FileSystemInfo defines the filesystem type and mountpoint of the device if it exists
type FileSystemInfo struct {
	// Type represents the FileSystem type of the block device
	// +optional
	Type string `json:"fsType,omitempty"`

	//MountPoint represents the mountpoint of the block device.
	// +optional
	Mountpoint string `json:"mountPoint,omitempty"`
}

// DeviceDevLink holds the mapping between type and links like by-id type or by-path type link
type DeviceDevLink struct {
	// Kind is the type of link like by-id or by-path.
	// +kubebuilder:validation:Enum:=by-id;by-path
	Kind string `json:"kind,omitempty"`

	// Links are the soft links
	Links []string `json:"links,omitempty"`
}

// BlockDeviceStatus defines the observed state of BlockDevice
type BlockDeviceStatus struct {
	// ClaimState represents the claim state of the block device
	// +kubebuilder:validation:Enum:=Claimed;Unclaimed;Released
	ClaimState DeviceClaimState `json:"claimState"`

	// State is the current state of the blockdevice (Active/Inactive/Unknown)
	// +kubebuilder:validation:Enum:=Active;Inactive;Unknown
	State BlockDeviceState `json:"state"`
}

// DeviceClaimState defines the observed state of BlockDevice
type DeviceClaimState string

const (
	// BlockDeviceUnclaimed represents that the block device is not bound to any BDC,
	// all cleanup jobs have been completed and is available for claiming.
	BlockDeviceUnclaimed DeviceClaimState = "Unclaimed"

	// BlockDeviceReleased represents that the block device is released from the BDC,
	// pending cleanup jobs
	BlockDeviceReleased DeviceClaimState = "Released"

	// BlockDeviceClaimed represents that the block device is bound to a BDC
	BlockDeviceClaimed DeviceClaimState = "Claimed"
)

// BlockDeviceState defines the observed state of the disk
type BlockDeviceState string

const (
	// BlockDeviceActive is the state for a block device that is connected to the node
	BlockDeviceActive BlockDeviceState = "Active"

	// BlockDeviceInactive is the state for a block device that is disconnected from a node
	BlockDeviceInactive BlockDeviceState = "Inactive"

	// BlockDeviceUnknown is the state for a block device whose state (attached/detached) cannot
	// be determined at this time.
	BlockDeviceUnknown BlockDeviceState = "Unknown"
)

// +kubebuilder:object:root=true

// BlockDeviceList contains a list of BlockDevice
type BlockDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlockDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BlockDevice{}, &BlockDeviceList{})
}
