package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepoolclaim

// StoragePoolClaim describes a StoragePoolClaim.
type StoragePoolClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StoragePoolClaimSpec `json:"spec"`
}

// StoragePoolClaimSpec is the spec for a StoragePoolClaimSpec resource
type StoragePoolClaimSpec struct {
	Name       string `json:"name"`       // disk name (e.g. sdb, sdc)
	Format     string `json:"format"`     // filesystem format (e.g. ext3, ext4)
	Mountpoint string `json:"mountpoint"` // where to mount the filesystem
	Path       string `json:"path"`       // disk path
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepoolclaims

// StoragePoolClaimList is a list of StoragePoolClaim resources
type StoragePoolClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePoolClaim `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepool

// StoragePool describes a StoragePool.
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StoragePoolSpec `json:"spec"`
}

// StoragePoolSpec is the spec for a StoragePool resource
type StoragePoolSpec struct {
	Name       string `json:"name"`       // name of the resource
	Format     string `json:"format"`     // filesystem format type (e.g. ext4, ext3)
	Mountpoint string `json:"mountpoint"` // mount location
	Nodename   string `json:"nodename"`   // host name
	Message    string `json:"message"`    // description
	Path       string `json:"path"`       // disk path
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools

// StoragePoolList is a list of StoragePool resources
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disk

// Disk describes disk resource.
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata, omitempty"`

	Spec DiskSpec `json:"spec"`
}

// DiskSpec is the specification for the disk stored as CRD
type DiskSpec struct {
	Path       string `json:"path"`         //disk path (e.g. /dev/sdb)
	Capacity DiskCapacity `json:"capacity"` //capacity (e.g. size, used)
	Details DiskDetails `json:"details"`    //disk details (e.g. model, serial)
}

type DiskCapacity struct {
	Storage   uint64 `json:"storage"`       //disk size in byte
}

type DiskDetails struct {
	Model       string `json:"model"`       //disk model number
	Serial      string `json:"serial"`      //disk serial number
	Vendor      string `json:"vendor"`      //disk vendor
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disks

// DiskList is a list of Disk object resources
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Disk `json:"items"`
}

