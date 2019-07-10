/*

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SidecarInjectSpec defines the desired state of SidecarInject
type SidecarInjectSpec struct {
	Selector            map[string]string `json:"selector"`
	SidecarNum          int            `json:"sidecarNum"`
	SidecarTemplateConfigmapName string            `json:"sidecarConfigmap"`
}

// SidecarInjectStatus defines the observed state of SidecarInject
type SidecarInjectStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SidecarInject is the Schema for the sidecarinjects API
// +k8s:openapi-gen=true
type SidecarInject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SidecarInjectSpec   `json:"spec,omitempty"`
	Status SidecarInjectStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SidecarInjectList contains a list of SidecarInject
type SidecarInjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SidecarInject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SidecarInject{}, &SidecarInjectList{})
}
