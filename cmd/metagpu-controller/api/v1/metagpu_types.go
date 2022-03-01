package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetaDevice struct {
	Node        string `json:"node"`
	UUID        string `json:"uuid"`
	DeviceIndex uint   `json:"deviceIndex"`
	MetaGpus    uint   `json:"slices"`
}

type MetaResource struct {
	ResourceName string        `json:"resourceName"`
	MetaDevices  []*MetaDevice `json:"metaDevices"`
}

type MetaDeviceConfiguration struct {
	ID          string `json:"id"`
	DeviceIndex uint   `json:"deviceIndex"`
	Shareable   bool   `json:"shareable"`
	MetaGpus    uint   `json:"slices"`
}

type MetaGpuSpec struct {
	DiscoveryMode string        `json:"discoveryMode"`
	MetaResources []*MetaDevice `json:"metaResources"`
}

type MetaGpuStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Foo",type=string,JSONPath=`.spec.foo`

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
