// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	SyncUserConditionType          ConditionType = "sync-user"
	SyncGroupMemberConditionType   ConditionType = "sync-group-member"
	SyncGroupConditionType         ConditionType = "sync-group"
	SyncProjectConditionType       ConditionType = "sync-project"
	SyncProjectMemberConditionType ConditionType = "sync-project-member"
)

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
	URL          string `json:"url"`
	ApiServer    string `json:"apiserver"`
	ProviderType string `json:"providertype"`
}

type Application struct {
	ApplicationRef  *ApplicationRef  `json:"application_ref"`
	ApplicationSpec *ApplicationSpec `json:"application_spec"`
}

type UserPermissionStream struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              UserPermissionStreamSpec   `json:"spec"`
	Status            UserPermissionStreamStatus `json:"status"`
}

type UserPermissionStreamSpec struct {
	Source  *Application   `json:"source"`
	Targets []*Application `json:"targets"`
}

type UserPermissionStreamStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}
