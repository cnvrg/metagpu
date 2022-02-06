package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetaGpuSpec struct {
	Foo string `json:"foo,omitempty"`
}

type MetaGpuStatus struct {
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type MetaGpu struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetaGpuSpec   `json:"spec,omitempty"`
	Status MetaGpuStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type MetaGpuList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetaGpu `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetaGpu{}, &MetaGpuList{})
}
