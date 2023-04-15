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

package idp

import (
	"context"

	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
)

//go:generate mockgen -destination idp_mock.go -package idp -source idp_interface.go
type Idp interface {
	Kind() IdpKind
	SetName(name string)
	GetName() string
	SetApiServerUrl(url string)
	SetSecretProvider(provider *secret_provider.SecretProvider)
	GetStaticUserById(id string) (*schema.User, error)
	GetUsers(ctx context.Context) ([]*schema.User, error)
	GetGroups(ctx context.Context) ([]*schema.Group, error)
	GetProjects(ctx context.Context) ([]*schema.Project, error)
	GetAllGroupMembers(ctx context.Context, groups []*schema.Group, users []*schema.User) ([]*schema.GroupMember, error)
	GetGroupMembers(ctx context.Context, group *schema.Group, user *schema.User) ([]*schema.GroupMember, error)
	GetProjectMembers(ctx context.Context, project *schema.Project, user *schema.User) ([]*schema.ProjectMember, error)
}
