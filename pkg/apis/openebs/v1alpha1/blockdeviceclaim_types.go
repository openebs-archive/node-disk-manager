package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// DeviceClaimSpec defines the desired state of BlockDeviceClaim
type DeviceClaimSpec struct {
	Resources       DeviceClaimResources `json:"resources"`                    // the resources in the claim like Capacity, IOPS
	DeviceType      string               `json:"deviceType"`                   // DeviceType represents the type of drive like SSD, HDD etc.,
	HostName        string               `json:"hostName"`                     // Node name from where blockdevice has to be claimed.
	Details         DeviceClaimDetails   `json:"deviceClaimDetails,omitempty"` // Details of the device to be claimed
	BlockDeviceName string               `json:"blockDeviceName,omitempty"`    // BlockDeviceName is the reference to the block-device backing this claim
}

// DeviceClaimStatus defines the observed state of BlockDeviceClaim
type DeviceClaimStatus struct {
	Phase DeviceClaimPhase `json:"phase"`
}

// DeviceClaimPhase is a typed string for phase field of BlockDeviceClaim.
type DeviceClaimPhase string

// BlockDeviceClaim CR, when created pass through phases before it got some Devices Assigned.
// Given below table, have all phases which BlockDeviceClaim CR can go before it is marked done.
const (
	// BlockDeviceClaimStatusEmpty represents that the BlockDeviceClaim was just created.
	BlockDeviceClaimStatusEmpty DeviceClaimPhase = ""

	// BlockDeviceClaimStatusPending represents BlockDeviceClaim has not been assigned devices yet. Rather
	// search is going on for matching devices.
	BlockDeviceClaimStatusPending DeviceClaimPhase = "Pending"

	// BlockDeviceClaimStatusInvalidCapacity represents BlockDeviceClaim has invalid capacity request i.e. 0/-1
	BlockDeviceClaimStatusInvalidCapacity DeviceClaimPhase = "Invalid Capacity Request"

	// BlockDeviceClaimStatusDone represents BlockDeviceClaim has been assigned backing blockdevice and ready for use.
	BlockDeviceClaimStatusDone DeviceClaimPhase = "Bound"
)

// DeviceClaimResources defines the request by the claim, eg, Storage, IOPS
type DeviceClaimResources struct {
	// Requests describes the minimum resources required. eg: if storage resource of 10G is
	// requested minimum capacity of 10G should be available
	Requests v1.ResourceList `json:"requests"`
}

const (
	// ResourceStorage defines the storage required as v1.Quantity
	ResourceStorage v1.ResourceName = "storage"
)

// DeviceClaimDetails defines the details of the block device that should be claimed
type DeviceClaimDetails struct {
	// BlockVolumeMode represents whether to claim a device in Block mode or Filesystem mode.
	// These are use cases of BlockVolumeMode:
	// 1) Not specified: DeviceFormat and MountPoint will not be considered
	// 2) VolumeModeBlock: DeviceFormat and MountPoint checks will be used as empty strings irrespective
	//    of the value they hold
	// 3) VolumeModeFileSystem: DeviceFormat and MountPoint will be used for exact matches
	BlockVolumeMode BlockDeviceVolumeMode `json:"blockVolumeMode, omitempty"`
	DeviceFormat    string                `json:"formatType,omitempty"`     //Format of the device required, eg:ext4, xfs
	MountPoint      string                `json:"mountPoint,omitempty"`     //MountPoint of the device required. Claim device from the specified mountpoint.
	AllowPartition  bool                  `json:"allowPartition,omitempty"` //AllowPartition represents whether to claim a full block device or a device that is a partition
}

// BlockDeviceVolumeMode
type BlockDeviceVolumeMode string

const (
	VolumeModeBlock      BlockDeviceVolumeMode = "Block"
	VolumeModeFileSystem BlockDeviceVolumeMode = "FileSystem"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// BlockDeviceClaim is the Schema for the block device claim API
type BlockDeviceClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceClaimSpec   `json:"spec,omitempty"`
	Status DeviceClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlockDeviceClaimList contains a list of BlockDeviceClaim
type BlockDeviceClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlockDeviceClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BlockDeviceClaim{}, &BlockDeviceClaimList{})
}
