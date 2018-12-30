/*
Copyright 2018 Eric Kfir.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IpAddressSpec defines the desired state of IpAddress
type IpAddressSpec struct {
	// +kubebuilder:validation:Pattern=[a-z][a-z0-9-]{4,28}[a-z0-9]
	ProjectId string `json:"projectId"`
	// +kubebuilder:validation:Enum=PREMIUM,STANDARD
	NetworkTier string `json:"networkTier"`
	// +kubebuilder:validation:Enum=IPV4,IPV6
	IpVersion string `json:"ipVersion"`
	Region    string `json:"region,omitempty"`
}

// IpAddressStatus defines the observed state of IpAddress
type IpAddressStatus struct {
	Address string `json:"Address,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpAddress is the Schema for the ipaddresses API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Project ID",type="string",JSONPath=".spec.ProjectId",description="Project ID"
// +kubebuilder:printcolumn:name="Tier",type="string",JSONPath=".spec.NetworkTier",description="Network Tier"
// +kubebuilder:printcolumn:name="IP Version",type="string",JSONPath=".spec.IpVersion",description="IP version"
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.Region",description="Region"
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".status.Address",description="IP address"
type IpAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpAddressSpec   `json:"spec,omitempty"`
	Status IpAddressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpAddressList contains a list of IpAddress
type IpAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpAddress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpAddress{}, &IpAddressList{})
}
