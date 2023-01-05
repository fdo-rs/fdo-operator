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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FDOManufacturingSpec defines the desired state of FDOManufacturing
type FDOManufacturingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Desired number of replicas
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// Container image
	Image string `json:"image,omitempty"`

	// Resources allocated for a manufacturing server pod (e.g. CPU)
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Name of the storage class to use for ownership vouchers persistent volumes
	StorageClassName string `json:"storageClassName,omitempty"`

	// Hostname of the route the server will be exposed through
	RouteHost string `json:"routeHost,omitempty"`

	// Log level: TRACE, DEBUG, INFO(default), WARN, ERROR or OFF
	// +kubebuilder:validation:Enum=TRACE;DEBUG;INFO;WARN;ERROR;OFF
	LogLevel string `json:"logLevel,omitempty"`

	// List of rendezvous servers
	// +listType=atomic
	RendezvousServers []RendezvousServer `json:"rendezvousServers"`

	// TODO:
	Protocols *Protocols `json:"protocols"`
}

//RendezvousServer defines an entry of rendezvous server configuration
// TODO: Implement full configuration parameters of the reference implementation
type RendezvousServer struct {

	// Hostname of a rendezvous server, must select either a hostname or an IP address
	// TODO: Add validation
	DNS string `json:"dns,omitempty"`

	// IP address of a rendezvous server, must select either an IP address or a hostname
	// TODO: Add validation
	IPAddress string `json:"ipAddress,omitempty"`

	// Rendezvous port for device connections
	DevicePort uint16 `json:"devicePort,omitempty"`

	// Rendezvous port for owner connections
	OwnerPort uint16 `json:"ownerPort,omitempty"`

	// Rendezvous transport protocol - tcp, tls (default), http, coap, https or coaps
	// +kubebuilder:validation:Enum=tcp;tls;http;coap;https;coaps
	Protocol string `json:"protocol,omitempty"`
}

type Protocols struct {
	PlainDI bool  `json:"plainDI"`
	DIUN    *DIUN `json:"diun,omitempty"`
}

type DIUN struct {
	// +kubebuilder:validation:Enum=SECP256R1;SECP384R1
	KeyType string `json:"keyType"`
	// +kubebuilder:validation:Enum=FileSystem;Tpm
	// +kubebuilder:valdation:MinLength=1
	AllowedKeyStorageTypes []string `json:"allowedKeyStorageTypes"`
}

// FDOManufacturingStatus defines the observed state of FDOManufacturing
type FDOManufacturingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Pods []string `json:"pods,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FDOManufacturing is the Schema for the fdomanufacturings API
type FDOManufacturing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FDOManufacturingSpec   `json:"spec,omitempty"`
	Status FDOManufacturingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FDOManufacturingList contains a list of FDOManufacturing
type FDOManufacturingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FDOManufacturing `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FDOManufacturing{}, &FDOManufacturingList{})
}
