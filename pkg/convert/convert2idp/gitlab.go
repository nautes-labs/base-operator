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

package convert2idp

// read gitlab data convert to idp struct

import (
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/spf13/cast"
	"github.com/xanzy/go-gitlab"
)

type Gitlab2IdpConverter struct {
}

func NewGitlab2IdpConverter() *Gitlab2IdpConverter {
	return &Gitlab2IdpConverter{}
}

func (*Gitlab2IdpConverter) ToIdpUser(gitlabUser *gitlab.User) *schema.User {
	user := &schema.User{
		BaseEntity: schema.BaseEntity{
			Identity:    cast.ToString(gitlabUser.ID),
			Name:        gitlabUser.Name,
			Description: "",
		},
		Username:    gitlabUser.Username,
		Email:       gitlabUser.Email,
		AvatarURL:   gitlabUser.AvatarURL,
		Mobile:      "",
		NamespaceId: cast.ToString(gitlabUser.NamespaceID),
	}
	return user
}

func (*Gitlab2IdpConverter) ToIdpGroup(gitlabGroup *gitlab.Group, kind string) *schema.Group {
	group := &schema.Group{
		BaseEntity: schema.BaseEntity{
			Identity:    cast.ToString(gitlabGroup.ID),
			Name:        gitlabGroup.Name,
			Description: gitlabGroup.Description,
		},
		Kind:     kind,
		ChildIds: make([]string, 0),
	}
	if gitlabGroup.ParentID > 0 {
		group.ParentId = cast.ToString(gitlabGroup.ParentID)
	}
	return group
}

func (*Gitlab2IdpConverter) ToIdpProject(gitlabProject *gitlab.Project) *schema.Project {
	project := &schema.Project{
		BaseEntity: schema.BaseEntity{
			Identity:    cast.ToString(gitlabProject.ID),
			Name:        gitlabProject.Name,
			Description: gitlabProject.Description,
		},
		Namespace: &schema.ProjectNamespace{
			Identity: cast.ToString(gitlabProject.Namespace.ID),
			Kind:     gitlabProject.Namespace.Kind,
			ParentId: cast.ToString(gitlabProject.Namespace.ParentID),
		},
	}
	return project
}

func (*Gitlab2IdpConverter) ToIdpGroupMember(groupId string, gitlabGroupMember *gitlab.GroupMember) *schema.GroupMember {
	groupMember := &schema.GroupMember{
		UserId:  cast.ToString(gitlabGroupMember.ID),
		GroupId: groupId,
	}
	return groupMember
}

// func (*Gitlab2IdpConverter) ToIdpProjectMember(projectId string, gitlabProjectMember *gitlab.ProjectMember) *schema.ProjectMember {
// 	project := &schema.ProjectMember{
// 		Id:        fmt.Sprintf("%d-%s", gitlabProjectMember.ID, projectId),
// 		UserId:    cast.ToString(gitlabProjectMember.ID),
// 		ProjectId: projectId,
// 	}
// 	return project
// }
