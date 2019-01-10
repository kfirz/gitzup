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

type DnsRecord struct {
	Type    string   `json:"type"`
	DnsName string   `json:"dnsName"`
	Ttl     int64    `json:"ttl"`
	Rrdatas []string `json:"rrdatas"`
}

// GoogleCloudDnsZoneSpec defines the desired state of GoogleCloudDnsZone
type DnsZoneSpec struct {

	// +kubebuilder:validation:Pattern=[a-z][a-z0-9-]{4,28}[a-z0-9]
	ProjectId string `json:"projectId"`

	// +kubebuilder:validation:Pattern=[^.]+\.[^.]+\.
	DnsName string `json:"dnsName"`

	// +kubebuilder:validation:Pattern=[a-z][a-z0-9-]*[a-z0-9]
	ZoneName string `json:"zoneName,omitempty"`

	Records []DnsRecord `json:"records"`
}

// GoogleCloudDnsZoneStatus defines the observed state of GoogleCloudDnsZone
type DnsZoneStatus struct {
	Id uint64 `json:"Id,omitempty"`
	ZoneName string `json:"zoneName,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GoogleCloudDnsZone is the Schema for the Google Cloud DNS Zones API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Project ID",type="string",JSONPath=".spec.ProjectId",description="Project ID"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.Id",description="Cloud DNS Zone ID"
// +kubebuilder:printcolumn:name="DNS",type="string",JSONPath=".spec.DnsName",description="DNS"
type DnsZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DnsZoneSpec   `json:"spec,omitempty"`
	Status DnsZoneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DnsZoneList contains a list of DnsZone
type DnsZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DnsZone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DnsZone{}, &DnsZoneList{})
}
