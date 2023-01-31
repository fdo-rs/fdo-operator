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

// FDOOnboardingServerSpec defines the desired state of FDOOnboardingServer
type FDOOnboardingServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Owner-onboarding server container image
	// +kubebuilder:default="quay.io/vemporop/fdo-owner-onboarding-server:1.0"
	OwnerOnboardingImage string `json:"ownerOnboardingImage,omitempty"`

	// ServiceInfo API server container image
	// +kubebuilder:default="quay.io/vemporop/fdo-serviceinfo-api-server:1.0"
	ServiceInfoImage string `json:"serviceInfoImage,omitempty"`

	// Service info device onboarding sequence
	ServiceInfo *ServiceInfo `json:"serviceInfo,omitempty"`
}

// FDOOnboardingServerStatus defines the observed state of FDOOnboardingServer
type FDOOnboardingServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Pods lists all pods running the onboarding server
	Pods []string `json:"pods,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FDOOnboardingServer is the Schema for the fdoonboardingservers API
type FDOOnboardingServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FDOOnboardingServerSpec   `json:"spec,omitempty"`
	Status FDOOnboardingServerStatus `json:"status,omitempty"`
}

// ServiceInfo defines a custom device onboarding sequence run through service info API
type ServiceInfo struct {
	InitialUser            *InitialUser           `json:"initialUser,omitempty"`
	Commands               []Command              `json:"commands,omitempty"`
	DiskEncryptionClevises []DiskEncryptionClevis `json:"diskencryptionClevis,omitempty"`
}

type InitialUser struct {
	Username string   `json:"username"`
	SSHKeys  []string `json:"sshKeys"`
}

type Command struct {
	Command      string   `json:"command"`
	Args         []string `json:"args"`
	MayFail      bool     `json:"mayFail,omitempty"`
	ReturnStdOut bool     `json:"returnStdOut,omitempty"`
	ReturnStdErr bool     `json:"returnStdErr,omitempty"`
}

type DiskEncryptionClevis struct {
	DiskLabel string                                  `json:"diskLabel"`
	Binding   *ServiceInfoDiskEncryptionClevisBinding `json:"binding"`
	ReEncrypt bool                                    `json:"reencrypt"`
}

type ServiceInfoDiskEncryptionClevisBinding struct {
	Pin    string `json:"pin,omitempty"`
	Config string `json:"config,omitempty"`
}

func (m *FDOOnboardingServer) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *FDOOnboardingServer) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// FDOOnboardingServerList contains a list of FDOOnboardingServer
type FDOOnboardingServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FDOOnboardingServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FDOOnboardingServer{}, &FDOOnboardingServerList{})
}
