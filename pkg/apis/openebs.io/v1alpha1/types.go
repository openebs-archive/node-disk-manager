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
	Path     string       `json:"path"`     //disk path (e.g. /dev/sdb)
	Capacity DiskCapacity `json:"capacity"` //capacity (e.g. size, used)
	Details  DiskDetails  `json:"details"`  //disk details (e.g. model, serial)
}

type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

type DiskCapacity struct {
	Storage uint64 `json:"storage"` //disk size in byte
}

type DiskDetails struct {
	Model  string `json:"model"`  //disk model number
	Serial string `json:"serial"` //disk serial number
	Vendor string `json:"vendor"` //disk vendor
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disks

// DiskList is a list of Disk object resources
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Disk `json:"items"`
}
