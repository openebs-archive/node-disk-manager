package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DiskSpec defines the desired state of Disk
type DiskSpec struct {
	Path             string        `json:"path"`                       //Path contain devpath (e.g. /dev/sdb)
	Capacity         DiskCapacity  `json:"capacity"`                   //Capacity
	Details          DiskDetails   `json:"details"`                    //Details contains static attributes (model, serial ..)
	FileSystem       string        `json:"fileSystem,omitempty"`       //Contains the data about filesystem on the disk
	PartitionDetails []Partition   `json:"partitionDetails,omitempty"` //Details of partitions in the disk (filesystem, partition type)
	DevLinks         []DiskDevLink `json:"devlinks,omitempty"`         //DevLinks contains soft links of one disk
}

type DiskCapacity struct {
	Storage            uint64 `json:"storage"`             // disk size in bytes
	PhysicalSectorSize uint32 `json: "physicalSectorSize"` // disk physical-Sector size in bytes
	LogicalSectorSize  uint32 `json:"logicalSectorSize"`   // disk logical-sector size in bytes
}

type DiskDetails struct {
	RotationRate     uint16 `json: "rotationRate"`    // Disk rotation speed if disk is not SSD
	DriveType        string `json: "driveType"`       // DriveType represents the type of drive like SSD, HDD etc.,
	Model            string `json:"model"`            // Model is model of disk
	Compliance       string `json:"compliance"`       // Implemented standards/specifications version such as SPC-1, SPC-2, etc
	Serial           string `json:"serial"`           // Serial is serial no of disk
	Vendor           string `json:"vendor"`           // Vendor is vendor of disk
	FirmwareRevision string `json:"firmwareRevision"` // disk firmware revision
}

// DiskDevlink holds the maping between type and links like by-id type or by-path type link
type DiskDevLink struct {
	Kind  string   `json:"kind"`  // Kind is the type of link like by-id or by-path.
	Links []string `json:"links"` // Links are the soft links of Type type
}

// DiskStatus defines the observed state of Disk
type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

type Temperature struct {
	CurrentTemperature int16 `json:"currentTemperature"`
	HighestTemperature int16 `json:"highestTemperature"`
	LowestTemperature  int16 `json:"lowestTemperature"`
}

type DiskStat struct {
	TempInfo              Temperature `json:"diskTemperature"`
	TotalBytesRead        uint64      `json:"totalBytesRead"`
	TotalBytesWritten     uint64      `json:"totalBytesWritten"`
	DeviceUtilizationRate float64     `json:"deviceUtilizationRate"`
	PercentEnduranceUsed  float64     `json:"percentEnduranceUsed"`
}

type DeviceInfo struct {
	DeviceUID string `json: "Device UID"` //Cross reference to Device CR backed by this disk
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Disk is the Schema for the disks API
// +k8s:openapi-gen=true
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec,omitempty"`
	Status DiskStatus `json:"status,omitempty"`
	Stats  DiskStat   `json:"stats"`
	Device DeviceInfo `json: "deviceInfo"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DiskList contains a list of Disk
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Disk `json:"items"`
}

type Partition struct {
	PartitionType  string `json:"partitionType"`
	FileSystemType string `json:"fileSystemType"`
}

func init() {
	SchemeBuilder.Register(&Disk{}, &DiskList{})
}
