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

package schema

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	NamespaceGroup   = "group"
	NamespaceUser    = "user"
	NamespaceProject = "project"
)

type BaseEntity struct {
	Identity    string `json:"identity"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type User struct {
	BaseEntity
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	AvatarURL   string   `json:"avatar_url"`
	Mobile      string   `json:"mobile"`
	RoleIds     []string `json:"role_Ids"`
	NamespaceId string   `json:"namespace_id"`
}

type Group struct {
	BaseEntity
	Kind     string   `json:"kind"`
	ParentId string   `json:"parent_id"`
	ChildIds []string `json:"child_ids"`
}

type Project struct {
	BaseEntity
	Namespace *ProjectNamespace `json:"namespace"`
	//ParentId string `json:"parent_id"`
}

type ProjectNamespace struct {
	Identity string `json:"id"`
	Kind     string `json:"kind"`
	ParentId string `json:"parent_id"`
}

type GroupMember struct {
	Id      string `json:"id"`
	GroupId string `json:"group_id"`
	UserId  string `json:"user_id"`
}

type ProjectMember struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	ProjectId string `json:"project_id"`
}

type TargetKNRI struct {
	Kind     string
	Name     string
	RoleKind string
	Identity string
}

func (t TargetKNRI) IsEmpty() bool {
	if len(t.Kind) == 0 {
		return true
	}
	if len(t.Name) == 0 {
		return true
	}
	if len(t.RoleKind) == 0 {
		return true
	}
	if len(t.Identity) == 0 {
		return true
	}
	return false
}

func StringToKNRI(s string) TargetKNRI {
	emptyKNRI := TargetKNRI{}
	identityArr := strings.Split(s, "-")
	if len(identityArr) != 4 {
		return emptyKNRI
	}
	emptyKNRI.Kind = identityArr[0]
	emptyKNRI.Name = identityArr[1]
	emptyKNRI.RoleKind = identityArr[2]
	emptyKNRI.Identity = identityArr[3]
	return emptyKNRI
}

func UserIsChanged(old *User, new *User) bool {
	return !cmp.Equal(*old, *new, cmpopts.IgnoreFields(User{}, "Identity", "RoleIds", "NamespaceId"))
}

func GroupIsChanged(old *Group, new *Group) bool {
	return !cmp.Equal(*old, *new,
		cmpopts.IgnoreFields(Group{}, "Identity", "ParentId"),
		cmpopts.IgnoreSliceElements(func(item string) bool {
			if strings.Contains(item, "project") {
				return true
			}
			return false
		}))
}

func ProjectIsChanged(old *Project, new *Project) bool {
	return !cmp.Equal(*old, *new, cmpopts.IgnoreFields(Project{}, "Identity", "Namespace"))
}

func RenderChildGroupIds(groups []*Group) {
	mapping := make(map[string][]string, 0)
	for _, group := range groups {
		if len(group.ParentId) == 0 {
			continue
		}
		mapping[group.ParentId] = append(mapping[group.ParentId], group.Identity)
	}
	for _, group := range groups {
		if childIds, ok := mapping[group.Identity]; ok {
			group.ChildIds = childIds
		}
	}
	return
}

func GroupMembersToUsers(groupMembers []*GroupMember) []*User {
	users := make([]*User, 0)
	mapper := make(map[string][]string)
	for _, groupMember := range groupMembers {
		mapper[groupMember.UserId] = append(mapper[groupMember.UserId], groupMember.GroupId)
	}
	for userId, groupIds := range mapper {
		user := &User{
			BaseEntity: BaseEntity{
				Identity: userId,
			},
			RoleIds: groupIds,
		}
		users = append(users, user)
	}
	return users
}
