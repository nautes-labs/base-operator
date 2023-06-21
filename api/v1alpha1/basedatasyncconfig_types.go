/*
Copyright 2022.

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

const (
	SyncUserConditionType          ConditionType = "sync-user"
	SyncGroupMemberConditionType   ConditionType = "sync-group-member"
	SyncGroupConditionType         ConditionType = "sync-group"
	SyncProjectConditionType       ConditionType = "sync-project"
	SyncProjectMemberConditionType ConditionType = "sync-project-member"
)

// +kubebuilder:object:generate=false
type ConditionType string

func (c ConditionType) ToString() string {
	return string(c)
}

type ApplicationRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
}

type ApplicationSpec struct {
	Name         string `json:"name"`
	ApiServerUrl string `json:"apiServerUrl"`
	ProviderType string `json:"providerType"`
}

type Application struct {
	// +optional
	ApplicationRef *ApplicationRef `json:"applicationRef"`
	// +optional
	ApplicationSpec *ApplicationSpec `json:"applicationSpec"`
}

// BaseDataSyncConfigSpec defines the desired state of BaseDataSyncConfig
type BaseDataSyncConfigSpec struct {
	Source  *Application   `json:"source"`
	Targets []*Application `json:"targets"`
}

// BaseDataSyncConfigStatus defines the observed state of BaseDataSyncConfig
type BaseDataSyncConfigStatus struct {
	// +optional
	Conditions []metav1.Condition `json:"conditions"`
	// +optional
	TargetStatus map[string][]metav1.Condition `json:"targetStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName={base-cfg,basecfg}
//+kubebuilder:subresource:status

// BaseDataSyncConfig is the Schema for the basedatasyncconfigs API
type BaseDataSyncConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BaseDataSyncConfigSpec   `json:"spec,omitempty"`
	Status BaseDataSyncConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BaseDataSyncConfigList contains a list of BaseDataSyncConfig
type BaseDataSyncConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BaseDataSyncConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BaseDataSyncConfig{}, &BaseDataSyncConfigList{})
}
