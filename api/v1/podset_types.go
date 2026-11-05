package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Desired state of PodSet
type PodSetSpec struct {
	Replicas int32  `json:"replicas"`
	Image    string `json:"image,omitempty"`
}

// PodSetStatus defines the observed state of PodSet (Current State).
type PodSetStatus struct {
	Ready   bool   `json:"Ready"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PodSet is the Schema for the podsets API
type PodSet struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PodSet
	// +required
	Spec PodSetSpec `json:"spec"`

	// status defines the observed state of PodSet
	// +optional
	Status PodSetStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PodSetList contains a list of PodSet
type PodSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PodSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PodSet{}, &PodSetList{})
		return nil
	})
}
