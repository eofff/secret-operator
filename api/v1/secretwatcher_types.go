/*
Copyright 2024.

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

// SecretWatcherSpec defines the desired state of SecretWatcher
type SecretWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SecretWatcher. Edit secretwatcher_types.go to remove/update

	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:validation:Required
	CheckTimeoutSeconds int `json:"checkTimeoutSeconds"`

	// +kubebuilder:validation:Required
	Namespaces []string `json:"namespaces"`
}

// SecretWatcherStatus defines the observed state of SecretWatcher
type SecretWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecretWatcher is the Schema for the secretwatchers API
type SecretWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretWatcherSpec   `json:"spec,omitempty"`
	Status SecretWatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretWatcherList contains a list of SecretWatcher
type SecretWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretWatcher{}, &SecretWatcherList{})
}