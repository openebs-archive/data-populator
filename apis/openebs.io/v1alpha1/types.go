/*
Copyright © 2022 The OpenEBS Authors

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

// +kubebuilder:object:generate=true
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatusWaitingForConsumer = "WaitingForConsumer"
	StatusInProgress         = "InProgress"
	StatusCompleted          = "Completed"
	StatusFailed             = "Failed"
)

// RsyncPopulator is a volume populator that helps
// to create a volume from any rsync source.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RsyncPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RsyncPopulatorSpec `json:"spec"`
}

// RsyncPopulatorList is a list of RsyncPopulator objects
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RsyncPopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []RsyncPopulator `json:"items"`
}

// RsyncPopulatorSpec contains the information of rsync daemon.
type RsyncPopulatorSpec struct {
	// Username is used as credential to access rsync daemon by the client.
	Username string `json:"username"`
	// Password is used as credential to access rsync daemon by the client.
	Password string `json:"password"`
	// Path represent mount path of the volume which we want to sync by the client.
	Path string `json:"path"`
	// URL is rsync daemon url it can be dns can be ip:port. Client will use
	// it to connect and get the data from daemon.
	URL string `json:"url"`
}

// DataPopulator contains information used for populating volume from
// a given to a desired destination
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec contains details of rsync source/ rsync daemon. Rsync client will
	// use these information to get the data for the volume.
	Spec DataPopulatorSpec `json:"spec"`
	// +optional
	Status DataPopulatorStatus `json:"status"`
}

// DataPopulatorSpec contains information of the source and target pvc
type DataPopulatorSpec struct {
	// SourcePVC is name of the PVC that we want to copy data from
	SourcePVC string `json:"sourcePVC"`
	// SourcePVCNamespace is namespace of the PVC that we want to copy
	SourcePVCNamespace string `json:"sourcePVCNamespace"`
	// DestinationPVC is new PVC name. it will be created in openebs- namespace
	DestinationPVC corev1.PersistentVolumeClaimSpec `json:"destinationPVC"`
}

// DataPopulatorStatus contains status of volume copy
type DataPopulatorStatus struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// DataPopulatorList is a list of DataPopulator objects
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataPopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of VolumeCopies
	Items []DataPopulator `json:"items"`
}
