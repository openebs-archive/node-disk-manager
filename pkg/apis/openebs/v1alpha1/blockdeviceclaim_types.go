package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// DeviceClaimSpec defines the desired state of BlockDeviceClaim
type DeviceClaimSpec struct {
	Capacity   uint64             `json:"capacity"`           // disk size in bytes
	DeviceType string             `json:"deviceType"`         // DeviceType represents the type of drive like SSD, HDD etc.,
	HostName   string             `json:"hostName"`           // Node name from where blockdevice has to be claimed.
	Details    DeviceClaimDetails `json:"deviceClaimDetails"` // Details of the device to be claimed
}

// DeviceClaimPhase is a typed string for phase field of BlockDeviceClaim.
type DeviceClaimPhase string

/*
 * BlockDeviceClaim CR, when created pass through phases before it got some Devices Assigned.
 * Given below table, have all phases which BlockDeviceClaim CR can go before it is marked done.
 */
const (
	// BlockDeviceClaimStatusEmpty: BlockDeviceClaim CR is just created.
	BlockDeviceClaimStatusEmpty DeviceClaimPhase = ""

	// BlockDeviceClaimStatusPending: BlockDeviceClaim CR yet to be assigned devices. Rather
	// search is going on for matching devices.
	BlockDeviceClaimStatusPending DeviceClaimPhase = "Pending"

	// BlockDeviceClaimStatusInvalidCapacity:  BlockDeviceClaim CR has invalid capacity request i.e. 0/-1
	BlockDeviceClaimStatusInvalidCapacity DeviceClaimPhase = "Invalid Capacity Request"

	// BlockDeviceClaimStatusDone:  BlockDeviceClaim CR assigned backing blockdevice and ready for use.
	BlockDeviceClaimStatusDone DeviceClaimPhase = "Bound"
)

// DeviceClaimStatus defines the observed state of BlockDeviceClaim
type DeviceClaimStatus struct {
	Phase DeviceClaimPhase `json:"phase"`
}

type DeviceClaimDetails struct {
	DeviceFormat   string `json:"formatType"`     //Format of the device required, eg:ext4, xfs
	MountPoint     string `json:"mountPoint"`     //MountPoint of the device required. Claim device from the specified mountpoint.
	AllowPartition bool   `json:"allowPartition"` //AllowPartition represents whether to claim a full block device or a device that is a partition
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlockDeviceClaim is the Schema for the devicerequests API
// +k8s:openapi-gen=true
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
