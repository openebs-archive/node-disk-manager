package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// DeviceSpec defines the desired state of Device
type DeviceSpec struct {
	Path        string          `json:"path"`        //Path contain devpath (e.g. /dev/sdb)
	Capacity    DeviceCapacity  `json:"capacity"`    //Capacity
	Details     DeviceDetails   `json:"details"`     //Details contains static attributes (model, serial ..)
	DevLinks    []DeviceDevLink `json:"devlinks"`    //DevLinks contains soft links of one disk
	Partitioned string          `json:"partitioned"` //Device has partions or not (YES/NO)
}

type DeviceCapacity struct {
	Storage            uint64 `json:"storage"`             // device capacity in bytes
	PhysicalSectorSize uint32 `json: "physicalSectorSize"` // device physical-Sector size in bytes
	LogicalSectorSize  uint32 `json:"logicalSectorSize"`   // device logical-sector size in bytes
}

type DeviceDetails struct {
	DeviceType       string `json: "deviceType"`      // DeviceType represents the type of drive like SSD, HDD etc.,
	Model            string `json:"model"`            // Model is model of disk
	Compliance       string `json:"compliance"`       // Implemented standards/specifications version such as SPC-1, SPC-2, etc
	Serial           string `json:"serial"`           // Serial is serial no of disk
	Vendor           string `json:"vendor"`           // Vendor is vendor of disk
	FirmwareRevision string `json:"firmwareRevision"` // disk firmware revision
}

// DeviceDevlink holds the maping between type and links like by-id type or by-path type link
type DeviceDevLink struct {
	Kind  string   `json:"kind,omitempty"`  // Kind is the type of link like by-id or by-path.
	Links []string `json:"links,omitempty"` // Links are the soft links of Type type
}

// DeviceClaimState defines the observed state of Device
type DeviceClaimState struct {
	State string `json:"state"` //current claim state of the device (Claimed/Unclaimed)
}

// DeviceStatus defines the observed state of Device
type DeviceStatus struct {
	State string `json:"state"` //current state of the device (Active/Inactive)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Device is the Schema for the devices API
// +k8s:openapi-gen=true
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec       DeviceSpec       `json:"spec,omitempty"`
	Status     DeviceStatus     `json:"status,omitempty"`
	ClaimState DeviceClaimState `json:"claimState"`
	ClaimRef   *v1.ObjectReference
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceList contains a list of Device
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Device{}, &DeviceList{})
}
