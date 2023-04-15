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

import (
	"fmt"
	"strings"

	"github.com/nautes-labs/base-operator/pkg/nexus/schema/security"
	"github.com/nautes-labs/base-operator/pkg/schema"
)

type Nexus2IdpConverter struct {
}

func NewNexus2IdpConverter() *Nexus2IdpConverter {
	return &Nexus2IdpConverter{}
}

func (*Nexus2IdpConverter) ToIdpUser(nexusUser *security.User) *schema.User {
	user := &schema.User{
		BaseEntity: schema.BaseEntity{
			Identity:    nexusUser.UserID,
			Name:        nexusUser.LastName,
			Description: "",
		},
		Username:  nexusUser.FirstName,
		Email:     nexusUser.EmailAddress,
		AvatarURL: "",
		Mobile:    "",
		RoleIds:   nexusUser.Roles,
	}
	return user
}

func (*Nexus2IdpConverter) RoleToGroup(idpKind string, idpName string, role *security.Role) *schema.Group {
	knri := schema.StringToKNRI(role.ID)
	if knri.IsEmpty() {
		return nil
	}
	if knri.Kind != idpKind {
		return nil
	}
	if knri.Name != idpName {
		return nil
	}
	if knri.RoleKind != schema.NamespaceUser && knri.RoleKind != schema.NamespaceGroup {
		return nil
	}
	group := &schema.Group{
		BaseEntity: schema.BaseEntity{
			Identity:    role.ID,
			Name:        role.Name,
			Description: role.Description,
		},
		Kind:     knri.RoleKind,
		ChildIds: role.Roles,
	}
	return group
}

func (*Nexus2IdpConverter) RolesToGroups(idpKind string, idpName string, roles []*security.Role) []*schema.Group {
	//existParentRoleIdMapping := make(map[string]string, len(roles))
	result := make([]*schema.Group, 0, len(roles))
	for _, role := range roles {
		// filter only group role and user namespace
		knri := schema.StringToKNRI(role.ID)
		if knri.IsEmpty() {
			continue
		}
		if knri.Kind != idpKind {
			continue
		}
		if knri.Name != idpName {
			continue
		}
		if knri.RoleKind != schema.NamespaceUser && knri.RoleKind != schema.NamespaceGroup {
			continue
		}
		baseEnt := schema.BaseEntity{
			Identity:    role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
		g := &schema.Group{
			BaseEntity: baseEnt,
			Kind:       knri.RoleKind,
		}
		if role.Roles == nil {
			role.Roles = make([]string, 0)
		}
		g.ChildIds = role.Roles
		result = append(result, g)

		// for _, roleId := range role.Roles {
		// 	existParentRoleIdMapping[roleId] = role.ID
		// }
	}

	// for _, item := range result {
	// 	if roleId, ok := existParentRoleIdMapping[item.Identity]; ok {
	// 		item.ParentId = roleId
	// 	}
	// }
	return result
}

func (*Nexus2IdpConverter) ToNormalRoles(idpKind string, idpName string, roles []*security.Role) []*schema.Group {
	result := make([]*schema.Group, 0, len(roles))
	for _, role := range roles {
		knri := schema.StringToKNRI(role.ID)
		if knri.IsEmpty() {
			continue
		}
		if knri.Kind != idpKind {
			continue
		}
		if knri.Name != idpName {
			continue
		}
		baseEnt := schema.BaseEntity{
			Identity:    role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
		g := &schema.Group{
			BaseEntity: baseEnt,
			Kind:       knri.RoleKind,
		}
		if role.Roles == nil {
			role.Roles = make([]string, 0)
		}
		g.ChildIds = role.Roles
		result = append(result, g)
	}
	return result
}

func (r *Nexus2IdpConverter) RolesToProjects(idpKind string, idpName string, roles []*security.Role) []*schema.Project {
	groups := r.RolesToGroups(idpKind, idpName, roles)
	mapping := make(map[string]string, len(groups))
	for _, group := range groups {
		for _, childId := range group.ChildIds {
			mapping[childId] = group.Identity
		}
	}
	result := make([]*schema.Project, 0, len(roles))
	for _, role := range roles {
		projectKnri := schema.StringToKNRI(role.ID)
		if projectKnri.IsEmpty() {
			continue
		}
		if projectKnri.Kind != idpKind {
			continue
		}
		if projectKnri.Name != idpName {
			continue
		}
		if projectKnri.RoleKind != schema.NamespaceProject {
			continue
		}
		item := &schema.Project{
			BaseEntity: schema.BaseEntity{
				Identity:    role.ID,
				Name:        role.Name,
				Description: role.Description,
			},
		}
		if gId, ok := mapping[role.ID]; ok {
			groupKnri := schema.StringToKNRI(gId)
			if groupKnri.IsEmpty() {
				continue
			}
			item.Namespace = &schema.ProjectNamespace{
				Identity: gId,
				Kind:     groupKnri.RoleKind,
			}
			if pGid, ok := mapping[gId]; ok {
				item.Namespace.ParentId = pGid
			}
		}
		result = append(result, item)
	}

	return result
}

func (*Nexus2IdpConverter) ToIdpProject(roles []*security.Role) []*schema.Project {
	result := make([]*schema.Project, 0, len(roles))
	mapping := make(map[string]string, len(roles))
	for _, role := range roles {
		for _, roleId := range role.Roles {
			mapping[roleId] = role.ID
		}
	}
	for _, role := range roles {
		item := &schema.Project{
			BaseEntity: schema.BaseEntity{
				Identity:    role.ID,
				Name:        role.Name,
				Description: role.Description,
			},
		}
		if gId, ok := mapping[role.ID]; ok {
			item.Namespace = &schema.ProjectNamespace{
				Identity: gId,
				Kind:     "group",
			}
			if pGid, ok := mapping[gId]; ok {
				item.Namespace.ParentId = pGid
			}
		}
		result = append(result, item)
	}
	return result
}

func (*Nexus2IdpConverter) ToIdpGroupMember(nexusUser *security.User) []*schema.GroupMember {
	groupMembers := make([]*schema.GroupMember, 0)
	for _, roleId := range nexusUser.Roles {
		if !strings.Contains(roleId, "group") {
			continue
		}
		groupMember := &schema.GroupMember{
			Id:      fmt.Sprintf("%s-%s", roleId, nexusUser.UserID),
			UserId:  nexusUser.UserID,
			GroupId: roleId,
		}
		groupMembers = append(groupMembers, groupMember)
	}
	return groupMembers
}

func (*Nexus2IdpConverter) ToIdpProjectMember(nexusUser *security.User) []*schema.ProjectMember {
	projectMembers := make([]*schema.ProjectMember, 0)
	for _, roleId := range nexusUser.Roles {
		if !strings.Contains(roleId, "project") {
			continue
		}
		projectMember := &schema.ProjectMember{
			Id:        fmt.Sprintf("%s-%s", roleId, nexusUser.UserID),
			UserId:    nexusUser.UserID,
			ProjectId: roleId,
		}
		projectMembers = append(projectMembers, projectMember)
	}
	return projectMembers
}
