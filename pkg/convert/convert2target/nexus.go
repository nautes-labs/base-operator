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

package convert2target

import (
	"github.com/nautes-labs/base-operator/pkg/nexus/schema/security"
	"github.com/nautes-labs/base-operator/pkg/schema"
)

type Idp2NexusConverter struct {
}

func NewIdp2NexusConverter() *Idp2NexusConverter {
	return &Idp2NexusConverter{}
}

func (*Idp2NexusConverter) IdpUser2NexusUser(identity string, user *schema.User, roleIds []string) *security.User {
	u := &security.User{
		UserID:       identity,
		FirstName:    user.Username,
		LastName:     user.Name,
		EmailAddress: user.Email,
	}
	if len(roleIds) > 0 {
		u.Roles = roleIds
	}
	return u
}

func (*Idp2NexusConverter) IdpGroup2NexusRole(identity string, group *schema.Group) *security.Role {
	role := &security.Role{
		ID:          identity,
		Name:        group.Name,
		Description: group.Description,
	}
	return role
}

func (*Idp2NexusConverter) IdpProject2NexusRole(identity string, project *schema.Project) *security.Role {
	role := &security.Role{
		ID:          identity,
		Name:        project.Name,
		Description: project.Description,
	}
	return role
}

// func IdpGroups2NexusRoles(groups []*schema.Group) []*security.Role {
// 	nexusRoles := make([]*security.Role, 0, len(groups))
// 	mapping := make(map[string][]string, 0)
// 	for _, group := range groups {
// 		if len(group.ParentId) == 0 {
// 			continue
// 		}
// 		mapping[group.ParentId] = append(mapping[group.ParentId], group.Id)
// 	}
// 	for _, group := range groups {
// 		item := &security.Role{
// 			ID:   group.Id,
// 			Name: group.Name,
// 		}
// 		if roleIds, ok := mapping[group.Id]; ok {
// 			item.Roles = roleIds
// 		}
// 		nexusRoles = append(nexusRoles, item)
// 	}
// 	return nexusRoles
// }

// func IdpProject2NexusRole(project *schema.Project) *security.Role {
// 	item := &security.Role{
// 		ID:   project.Id,
// 		Name: project.Name,
// 	}
// 	return item
// }
