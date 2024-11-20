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

// MysqlbackupSpec defines the desired state of Mysqlbackup.
type MysqlbackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Mysqlbackup. Edit mysqlbackup_types.go to remove/update
	Mysqlsvcname   string `json:"mysqlsvcname,omitempty"`
	Backupuser     string `json:"backupuser,omitempty"`
	Backuppassword string `json:"backuppassword,omitempty"`
	Backupdb       string `json:"backupdb,omitempty"`
	Cronjob        string `json:"cronjob,omitempty"`
}

// MysqlbackupStatus defines the observed state of Mysqlbackup.
type MysqlbackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Mysqlbackup is the Schema for the mysqlbackups API.
type Mysqlbackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlbackupSpec   `json:"spec,omitempty"`
	Status MysqlbackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MysqlbackupList contains a list of Mysqlbackup.
type MysqlbackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysqlbackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mysqlbackup{}, &MysqlbackupList{})
}
