package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NdmConfigSpec defines the desired state of NdmConfig
type NdmConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

type NdmConfigPhase string

const (
	NdmConfigPhaseInit   NdmConfigPhase = "Initializing"
	NdmConfigPhaseDone   NdmConfigPhase = "Init_Done"
	NdmConfigPhaseDelete NdmConfigPhase = "Delete"
)

// NdmConfigStatus defines the observed state of NdmConfig
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
type NdmConfigStatus struct {
	Phase NdmConfigPhase `json:"phase"` //Current state of NdmConfig (Init/Done/Delete)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NdmConfig is the Schema for the ndmconfigs API
// +k8s:openapi-gen=true
type NdmConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NdmConfigSpec   `json:"spec,omitempty"`
	Status NdmConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NdmConfigList contains a list of NdmConfig
type NdmConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NdmConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NdmConfig{}, &NdmConfigList{})
}
