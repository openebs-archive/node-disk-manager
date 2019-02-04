package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disk

// Disk describes disk resource.
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec"`
	Status DiskStatus `json:"status"`
	Stats  DiskStat   `json:"stats"`
}

// DiskSpec is the specification for the disk stored as CRD
type DiskSpec struct {
	Path             string        `json:"path"`                       //Path contain devpath (e.g. /dev/sdb)
	Capacity         DiskCapacity  `json:"capacity"`                   //Capacity
	Details          DiskDetails   `json:"details"`                    //Details contains static attributes (model, serial ..)
	FileSystem       string        `json:"fileSystem,omitempty"`       //Contains the data about filesystem on the disk
	PartitionDetails []Partition   `json:"partitionDetails,omitempty"` //Details of partitions in the disk (filesystem, partition type)
	DevLinks         []DiskDevLink `json:"devlinks,omitempty"`         //DevLinks contains soft links of one disk
}

type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

type DiskCapacity struct {
	Storage            uint64 `json:"storage"`            // disk size in bytes
	PhysicalSectorSize uint32 `json:"physicalSectorSize"` // disk physical-Sector size in bytes
	LogicalSectorSize  uint32 `json:"logicalSectorSize"`  // disk logical-sector size in bytes
}

// DiskDetails contains basic and static info of a disk
type DiskDetails struct {
	RotationRate     uint16 `json:"rotationRate"`     // Disk rotation speed if disk is not SSD
	DriveType        string `json:"driveType"`        // DriveType represents the type of drive like SSD, HDD etc.,
	Model            string `json:"model"`            // Model is model of disk
	Compliance       string `json:"compliance"`       // Implemented standards/specifications version such as SPC-1, SPC-2, etc
	Serial           string `json:"serial"`           // Serial is serial no of disk
	Vendor           string `json:"vendor"`           // Vendor is vendor of disk
	FirmwareRevision string `json:"firmwareRevision"` // disk firmware revision
}

// DiskDevlink holds the maping between type and links like by-id type or by-path type link
type DiskDevLink struct {
	Kind  string   `json:"kind,omitempty"`  // Kind is the type of link like by-id or by-path.
	Links []string `json:"links,omitempty"` // Links are the soft links of Type type
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disks

// DiskList is a list of Disk object resources
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Disk `json:"items"`
}

type DiskStat struct {
	TempInfo              Temperature `json:"diskTemperature"`
	TotalBytesRead        uint64      `json:"totalBytesRead"`
	TotalBytesWritten     uint64      `json:"totalBytesWritten"`
	DeviceUtilizationRate float64     `json:"deviceUtilizationRate"`
	PercentEnduranceUsed  float64     `json:"percentEnduranceUsed"`
}

type Temperature struct {
	CurrentTemperature int16 `json:"currentTemperature"`
	HighestTemperature int16 `json:"highestTemperature"`
	LowestTemperature  int16 `json:"lowestTemperature"`
}

type Partition struct {
	PartitionType  string `json:"partitionType"`
	FileSystemType string `json:"fileSystemType"`
}
