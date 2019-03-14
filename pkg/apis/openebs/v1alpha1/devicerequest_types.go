package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// DeviceRequestSpec defines the desired state of DeviceRequest
type DeviceRequestSpec struct {
	Capacity   uint64 `json:"capacity"`   // disk size in bytes
	DeviceType string `json:"deviceType"` // DeviceType represents the type of drive like SSD, HDD etc.,
	HostName   string `json:"hostName"`   // Node name from where device has to be claimed.
}

// DeviceRequestPhase is a typed string for phase field of DeviceRequest.
type DeviceRequestPhase string

/*
 * DeviceRequest CR, when created pass through phases before it got some Devices Assigned.
 * Given below table, have all phases which DeviceRequest CR can go before it is marked done.
 */
const (
	// DeviceRequestStatusEmpty: DeviceRequest CR is just created.
	DeviceRequestStatusEmpty DeviceRequestPhase = ""

	// DeviceRequestStatusPending: DeviceRequest CR yet to be assigned devices. Rather
	// search is going on for matching devices.
	DeviceRequestStatusPending DeviceRequestPhase = "Pending"

	// DeviceRequestStatusInvalidCapacity:  DeviceRequest CR has invalid capacity request i.e. 0/-1
	DeviceRequestStatusInvalidCapacity DeviceRequestPhase = "Invalid Capacity Request"

	// DeviceRequestStatusDone:  DeviceRequest CR assigned backing device and ready for use.
	DeviceRequestStatusDone DeviceRequestPhase = "Bound"
)

// DeviceRequestStatus defines the observed state of DeviceRequest
type DeviceRequestStatus struct {
	Phase DeviceRequestPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceRequest is the Schema for the devicerequests API
// +k8s:openapi-gen=true
type DeviceRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceRequestSpec   `json:"spec,omitempty"`
	Status DeviceRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceRequestList contains a list of DeviceRequest
type DeviceRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceRequest{}, &DeviceRequestList{})
}
