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

package target

import (
	"context"
	"fmt"

	"github.com/nautes-labs/base-operator/pkg/idp"
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
)

//go:generate mockgen -destination target_mock.go -package target -source target_interface.go
type TargetApp interface {
	IdentityKey() TargetAppKindName
	Kind() TargetAppKind
	SetIdp(idp.Idp)
	SetName(name string)
	GetName() string
	SetApiServerUrl(url string)
	SetSecretProvider(provider *secret_provider.SecretProvider)
	GetUsers(ctx context.Context) ([]*schema.User, error)
	GetGroups(ctx context.Context) ([]*schema.Group, error)
	GetProjects(ctx context.Context) ([]*schema.Project, error)
	GetGroupMembers(ctx context.Context) ([]*schema.GroupMember, error)
	GetProjectMembers(ctx context.Context, project *schema.Project, user *schema.User) ([]*schema.ProjectMember, error)
	CreateUser(ctx context.Context, user *schema.User) error
	UpdateUser(ctx context.Context, id string, user *schema.User) error
	CreateGroup(ctx context.Context, group *schema.Group) error
	WrappingUpAfterGroupSync(ctx context.Context) error
	UpdateGroup(ctx context.Context, id string, group *schema.Group) error
	CreateProject(ctx context.Context, project *schema.Project) error
	UpdateProject(ctx context.Context, id string, project *schema.Project) error
	CreateGroupMember(ctx context.Context, groupMember *schema.GroupMember) error
	UpdateGroupMember(ctx context.Context, id string, groupMember *schema.GroupMember) error
	CreateProjectMember(ctx context.Context, projectMember *schema.ProjectMember) error
	UpdateProjectMember(ctx context.Context, id string, projectMember *schema.ProjectMember) error
	DeleteUserById(ctx context.Context, id string) error
	DeleteGroupById(ctx context.Context, id string) error
	DeleteProjectById(ctx context.Context, id string) error
	DeleteGroupMemberById(ctx context.Context, id string) error
	DeleteProjectMemberById(ctx context.Context, id string) error
	GenerateIdpUserIdentity(Identity string) (idpUserIdentity string)
	GenerateIdpGroupIdentity(groupKind string, Identity string) (idpGroupIdentity string)
	GenerateIdpProjectIdentity(Identity string) (idpProjectIdentity string)
	CompareUsers(idpUsers []*schema.User, targetAppUsers []*schema.User) (createUsers []*schema.User, updateUsers []*schema.User)
	CompareGroups(idpGroups []*schema.Group, targetAppGroups []*schema.Group) (createGroups []*schema.Group, updateGroups []*schema.Group)
	CompareProjects(idpProjects []*schema.Project, targetAppProjects []*schema.Project) (createProjects []*schema.Project, updateProjects []*schema.Project)
	SyncGroupMember(ctx context.Context, idpGroupMembers []*schema.GroupMember, targetAppGroupMembers []*schema.GroupMember) error
	GroupBindingProjects(ctx context.Context, projects []*schema.Project) error
}

type TargetAppKindName struct {
	Kind string
	Name string
}

func (t TargetAppKindName) ToString() string {
	return fmt.Sprintf("%s-%s", t.Kind, t.Name)
}
