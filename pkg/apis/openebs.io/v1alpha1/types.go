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
	metav1.ObjectMeta `json:"metadata, omitempty"`

	Spec   DiskSpec   `json:"spec"`
	Status DiskStatus `json:"status"`
}

// DiskSpec is the specification for the disk stored as CRD
type DiskSpec struct {
	Path     string        `json:"path"`               //Path contain devpath (e.g. /dev/sdb)
	Capacity DiskCapacity  `json:"capacity"`           //Capacity
	Details  DiskDetails   `json:"details"`            //Details contains static attributes (model, serial ..)
	DevLinks []DiskDevLink `json:"devlinks,omitempty"` //DevLinks contains soft links of one disk
}

type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

type DiskCapacity struct {
	Storage           uint64 `json:"storage"`           // disk size in bytes
	LogicalSectorSize uint32 `json:"logicalSectorSize"` // disk logical size in bytes
}

// DiskDetails contains basic and static info of a disk
type DiskDetails struct {
	Model            string `json:"model"`            // Model is model of disk
	SPCVersion       string `json:"spcVersion"`       // Implemented standards/specifications version such as SPC-1, SPC-2, etc
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
