/*
Copyright 2021.

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

// DBaaSConfigSpec defines the desired state of DBaaSConfig
type DBaaSConfigSpec struct {
	// Defaults for inventories in this namespace
	DBaaSInventoryConfigs `json:",inline"`
}

// DBaaSConfigStatus defines the observed state of DBaaSConfig
type DBaaSConfigStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Active",type=string,JSONPath=`.status.conditions[0].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

//+operator-sdk:csv:customresourcedefinitions:displayName="DBaaSConfig"
// DBaaSConfig is the Schema for the dbaasconfigs API
type DBaaSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSConfigSpec   `json:"spec,omitempty"`
	Status DBaaSConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSConfigList contains a list of DBaaSConfig
type DBaaSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSConfig{}, &DBaaSConfigList{})
}
