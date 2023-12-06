/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FDORendezvousServerSpec defines the desired state of FDORendezvousServer
type FDORendezvousServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Rendezvous server container image
	// +kubebuilder:default="quay.io/vemporop/fdo-rendezvous-server:rhel9.3"
	Image string `json:"image,omitempty"`
}

// FDORendezvousServerStatus defines the observed state of FDORendezvousServer
type FDORendezvousServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Pods lists all pods running the rendezvous server
	Pods []string `json:"pods,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FDORendezvousServer is the Schema for the fdorendezvousservers API
type FDORendezvousServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FDORendezvousServerSpec   `json:"spec,omitempty"`
	Status FDORendezvousServerStatus `json:"status,omitempty"`
}

func (m *FDORendezvousServer) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *FDORendezvousServer) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// FDORendezvousServerList contains a list of FDORendezvousServer
type FDORendezvousServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FDORendezvousServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FDORendezvousServer{}, &FDORendezvousServerList{})
}
