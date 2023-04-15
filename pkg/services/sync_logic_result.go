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

package services

import "github.com/nautes-labs/base-operator/pkg/target"

const (
	SyncStatusSuccess = "True"
	SyncStatusFail    = "False"
)

const (
	ReadResourceKind    = "ReadResource"
	SyncUserKind        = "SyncUser"
	SyncGroupKind       = "SyncGroup"
	SyncProjectKind     = "SyncProject"
	SyncGroupMemberKind = "SyncGroupMember"
)

type SyncLogicResultItem struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

type SyncLogicResult struct {
	Brief  []*SyncLogicResultItem
	Detail map[target.TargetAppKindName][]*SyncLogicResultItem
}

func (r *SyncLogicResult) addBrief(items ...*SyncLogicResultItem) {
	r.Brief = append(r.Brief, items...)
}

func (r *SyncLogicResult) addDetail(instanceIdentity target.TargetAppKindName, items ...*SyncLogicResultItem) {
	r.Detail[instanceIdentity] = append(r.Detail[instanceIdentity], items...)
}

func NewReadResourceFailItem(msg string) *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    ReadResourceKind,
		Status:  SyncStatusFail,
		Reason:  SyncStatusFail,
		Message: msg,
	}
	return item
}

func NewReadResourceSuccessItem() *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    ReadResourceKind,
		Status:  SyncStatusSuccess,
		Reason:  SyncStatusSuccess,
		Message: "",
	}
	return item
}

func NewSyncUserFailItem(msg string) *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncUserKind,
		Status:  SyncStatusFail,
		Reason:  SyncStatusFail,
		Message: msg,
	}
	return item
}

func NewSyncUserSuccessItem() *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncUserKind,
		Status:  SyncStatusSuccess,
		Reason:  SyncStatusSuccess,
		Message: "",
	}
	return item
}

func NewSyncGroupFailItem(msg string) *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncGroupKind,
		Status:  SyncStatusFail,
		Reason:  SyncStatusFail,
		Message: msg,
	}
	return item
}

func NewSyncGroupSuccessItem() *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncGroupKind,
		Status:  SyncStatusSuccess,
		Reason:  SyncStatusSuccess,
		Message: "",
	}
	return item
}

func NewSyncProjectFailItem(msg string) *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncProjectKind,
		Status:  SyncStatusFail,
		Reason:  SyncStatusFail,
		Message: msg,
	}
	return item
}

func NewSyncProjectSuccessItem() *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncProjectKind,
		Status:  SyncStatusSuccess,
		Reason:  SyncStatusSuccess,
		Message: "",
	}
	return item
}

func NewSyncGroupMemberFailItem(msg string) *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncGroupMemberKind,
		Status:  SyncStatusFail,
		Reason:  SyncStatusFail,
		Message: msg,
	}
	return item
}

func NewSyncGroupMemberSuccessItem() *SyncLogicResultItem {
	item := &SyncLogicResultItem{
		Type:    SyncGroupMemberKind,
		Status:  SyncStatusSuccess,
		Reason:  SyncStatusSuccess,
		Message: "",
	}
	return item
}
