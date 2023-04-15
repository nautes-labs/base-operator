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
	"os"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nautes-labs/base-operator/pkg/convert/convert2idp"
	"github.com/nautes-labs/base-operator/pkg/convert/convert2target"
	"github.com/nautes-labs/base-operator/pkg/idp"
	"github.com/nautes-labs/base-operator/pkg/log"
	"github.com/nautes-labs/base-operator/pkg/nexus"
	"github.com/nautes-labs/base-operator/pkg/nexus/pkg/client"
	"github.com/nautes-labs/base-operator/pkg/nexus/schema/security"
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
	"github.com/nautes-labs/base-operator/pkg/util"
	"github.com/sirupsen/logrus"
)

var _ TargetApp = (*nexusApp)(nil)

type nexusApp struct {
	idp                idp.Idp
	name               string
	apiServerUrl       string
	client             *nexus.NexusClient
	secretProvider     *secret_provider.SecretProvider
	nexus2IdpConverter *convert2idp.Nexus2IdpConverter
	idp2NexusConverter *convert2target.Idp2NexusConverter
	roles              []*security.Role
	groups             []*schema.Group
}

func (n *nexusApp) newClient() error {
	if n.client == nil {
		username, passwd, err := n.secretProvider.GetApplicationBasicAuth(secret_provider.Identity{Type: string(n.Kind()), Name: n.GetName()})
		if err != nil {
			return fmt.Errorf("get basic auth info fail, err:%w", err)
		}
		//
		n.client = nexus.NewClient(client.Config{
			URL:      n.apiServerUrl,
			Username: username,
			Password: passwd,
			Insecure: true,
		})
	}
	return nil
}

func (n *nexusApp) IdentityKey() TargetAppKindName {
	return TargetAppKindName{
		Kind: string(n.Kind()),
		Name: n.GetName(),
	}
}

func (n *nexusApp) Kind() TargetAppKind {
	return NexusAppKind
}

func (n *nexusApp) SetIdp(idpEntity idp.Idp) {
	n.idp = idpEntity
}

func (n *nexusApp) SetName(name string) {
	n.name = name
	return
}

func (n *nexusApp) GetName() string {
	return n.name
}

func (n *nexusApp) SetApiServerUrl(url string) {
	n.apiServerUrl = url
	return
}

func (n *nexusApp) SetSecretProvider(provider *secret_provider.SecretProvider) {
	n.secretProvider = provider
	return
}

func (n *nexusApp) GetUsers(ctx context.Context) ([]*schema.User, error) {
	err := n.newClient()
	if err != nil {
		return nil, err
	}
	list, err := n.client.Security.User.List()
	if err != nil {
		return nil, err
	}
	result := make([]*schema.User, 0, len(list))
	for _, v := range list {
		result = append(result, n.nexus2IdpConverter.ToIdpUser(v))
	}
	return result, nil
}

func (n *nexusApp) getAllRoles() error {
	if len(n.roles) > 0 {
		return nil
	}
	err := n.newClient()
	if err != nil {
		return err
	}
	list, err := n.client.Security.Role.List()
	if err != nil {
		return err
	}
	n.roles = list
	return nil
}

func (n *nexusApp) GetGroups(ctx context.Context) ([]*schema.Group, error) {
	if len(n.groups) > 0 {
		return n.groups, nil
	}
	// invoke nexus client query all roles
	err := n.getAllRoles()
	if err != nil {
		return nil, err
	}
	result := n.nexus2IdpConverter.RolesToGroups(n.idp.Kind().Tostring(), n.idp.GetName(), n.roles)
	n.groups = result
	return result, nil
}

func (n *nexusApp) getLatestGroups() ([]*schema.Group, error) {
	err := n.newClient()
	if err != nil {
		return nil, err
	}
	list, err := n.client.Security.Role.List()
	if err != nil {
		return nil, err
	}
	result := n.nexus2IdpConverter.RolesToGroups(n.idp.Kind().Tostring(), n.idp.GetName(), list)
	return result, nil
}

func (n *nexusApp) addGroup(group *schema.Group) {
	n.groups = append(n.groups, group)
}

func (n *nexusApp) updateGroupInMemery(newGroup *schema.Group) {
	for _, g := range n.groups {
		if g.Identity == n.GenerateIdpGroupIdentity(newGroup.Kind, newGroup.Identity) {
			g.Name = newGroup.Name
			g.Description = newGroup.Description
			g.Kind = newGroup.Kind
			g.ChildIds = newGroup.ChildIds
			break
		}
	}
	return
}

func (n *nexusApp) GetProjects(ctx context.Context) ([]*schema.Project, error) {
	err := n.newClient()
	if err != nil {
		return nil, err
	}
	list, err := n.client.Security.Role.List()
	if err != nil {
		return nil, err
	}
	result := n.nexus2IdpConverter.RolesToProjects(n.idp.Kind().Tostring(), n.idp.GetName(), list)
	return result, nil
}

func (n *nexusApp) GetGroupMembers(ctx context.Context) ([]*schema.GroupMember, error) {
	users, err := n.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*schema.GroupMember, 0)
	for _, user := range users {
		for _, roleId := range user.RoleIds {
			knri := schema.StringToKNRI(roleId)
			if knri.IsEmpty() {
				continue
			}
			if knri.Kind != string(n.idp.Kind()) {
				continue
			}
			if knri.Name != n.idp.GetName() {
				continue
			}
			if !util.InArray(knri.RoleKind, []string{schema.NamespaceGroup, schema.NamespaceUser}) {
				continue
			}
			groupMember := &schema.GroupMember{
				UserId:  user.Identity,
				GroupId: roleId,
			}
			result = append(result, groupMember)
		}
	}
	return result, nil
}

func (n *nexusApp) GetProjectMembers(ctx context.Context, project *schema.Project, user *schema.User) ([]*schema.ProjectMember, error) {
	return nil, nil
}

func (n *nexusApp) CreateUser(ctx context.Context, user *schema.User) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	group := &schema.Group{
		BaseEntity: schema.BaseEntity{
			Identity:    user.NamespaceId,
			Name:        user.Username,
			Description: user.Description,
		},
		Kind: schema.NamespaceUser,
	}
	//1. create user namespace
	err = n.CreateGroup(ctx, group)
	if err != nil {
		return err
	}
	identity := n.GenerateIdpUserIdentity(user.Identity)
	groupIdentity := n.GenerateIdpGroupIdentity(schema.NamespaceUser, user.NamespaceId)
	u := n.idp2NexusConverter.IdpUser2NexusUser(identity, user, []string{groupIdentity})
	//2. create user
	err = n.client.Security.User.Create(*u)
	if err != nil {
		return err
	}
	return nil
}

func (n *nexusApp) UpdateUser(ctx context.Context, id string, user *schema.User) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	id = n.GenerateIdpUserIdentity(id)
	u := n.idp2NexusConverter.IdpUser2NexusUser(id, user, user.RoleIds)
	err = n.client.Security.User.Update(id, *u)
	if err != nil {
		return err
	}
	return nil
}

func (n *nexusApp) CreateGroup(ctx context.Context, group *schema.Group) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	groupIdentity := ""
	if group.Kind == schema.NamespaceGroup {
		groupIdentity = n.GenerateIdpGroupIdentity(schema.NamespaceGroup, group.Identity)
	} else {
		groupIdentity = n.GenerateIdpGroupIdentity(schema.NamespaceUser, group.Identity)
	}
	role := n.idp2NexusConverter.IdpGroup2NexusRole(groupIdentity, group)
	// create nexus role
	err = n.client.Security.Role.Create(*role)
	if err != nil {
		return err
	}
	// if group.Kind == schema.NamespaceGroup {
	// 	// add group in memery
	// 	memGroup := &schema.Group{
	// 		BaseEntity: schema.BaseEntity{
	// 			Identity:    groupIdentity,
	// 			Name:        group.Name,
	// 			Description: group.Description,
	// 		},
	// 		Kind:     group.Kind,
	// 		ParentId: group.ParentId,
	// 	}
	// 	if len(group.ParentId) > 0 {
	// 		memGroup.ParentId = n.GenerateIdpGroupIdentity(schema.NamespaceGroup, group.ParentId)
	// 	}
	// 	for _, childId := range group.ChildIds {
	// 		memGroup.ChildIds = append(memGroup.ChildIds, n.GenerateIdpGroupIdentity(schema.NamespaceGroup, childId))
	// 	}
	// 	n.addGroup(memGroup)
	// }
	return nil
}

func (n *nexusApp) UpdateGroup(ctx context.Context, id string, group *schema.Group) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	//id = n.GenerateIdpGroupIdentity(schema.NamespaceGroup, id)
	role := n.idp2NexusConverter.IdpGroup2NexusRole(id, group)
	role.Roles = group.ChildIds
	err = n.client.Security.Role.Update(id, *role)
	if err != nil {
		return err
	}
	//n.updateGroupInMemery(group)
	return nil
}

func (n *nexusApp) CreateProject(ctx context.Context, project *schema.Project) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	projectIdentity := n.GenerateIdpProjectIdentity(project.Identity)
	role := n.idp2NexusConverter.IdpProject2NexusRole(projectIdentity, project)
	err = n.client.Security.Role.Create(*role)
	if err != nil {
		return err
	}
	return nil
}

func (n *nexusApp) UpdateProject(ctx context.Context, id string, project *schema.Project) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	id = n.GenerateIdpProjectIdentity(id)
	role := n.idp2NexusConverter.IdpProject2NexusRole(id, project)
	// update role
	err = n.client.Security.Role.Update(id, *role)
	if err != nil {
		return err
	}
	return nil
}

func (n *nexusApp) CreateGroupMember(ctx context.Context, groupMember *schema.GroupMember) error {
	return nil
}

func (n *nexusApp) UpdateGroupMember(ctx context.Context, id string, groupMember *schema.GroupMember) error {
	return nil
}

func (n *nexusApp) CreateProjectMember(ctx context.Context, projectMember *schema.ProjectMember) error {
	return nil
}

func (n *nexusApp) UpdateProjectMember(ctx context.Context, id string, projectMember *schema.ProjectMember) error {
	return nil
}

func (n *nexusApp) DeleteUserById(ctx context.Context, id string) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	err = n.client.Security.User.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (n *nexusApp) DeleteGroupById(ctx context.Context, id string) error {
	err := n.newClient()
	if err != nil {
		return err
	}
	err = n.client.Security.Role.Delete(id)
	if err != nil {
		return err
	}
	return nil
}
func (n *nexusApp) DeleteProjectById(ctx context.Context, id string) error {
	return nil
}

func (n *nexusApp) DeleteGroupMemberById(ctx context.Context, id string) error {
	return nil
}
func (n *nexusApp) DeleteProjectMemberById(ctx context.Context, id string) error {
	return nil
}

func (n *nexusApp) GenerateIdpUserIdentity(Identity string) (idpUserIdentity string) {
	return fmt.Sprintf("%s-%s-%s", n.idp.Kind(), n.idp.GetName(), Identity)
}

func (n *nexusApp) GenerateIdpGroupIdentity(groupKind string, Identity string) (idpGroupIdentity string) {
	return fmt.Sprintf("%s-%s-%s-%s", n.idp.Kind(), n.idp.GetName(), groupKind, Identity)
}

func (n *nexusApp) GenerateIdpProjectIdentity(Identity string) (idpProjectIdentity string) {
	return fmt.Sprintf("%s-%s-%s-%s", n.idp.Kind(), n.idp.GetName(), schema.NamespaceProject, Identity)
}

func (n *nexusApp) CompareUsers(idpUsers []*schema.User, targetAppUsers []*schema.User) (createUsers []*schema.User, updateUsers []*schema.User) {
	targetAppUserIds := make([]string, 0, len(targetAppUsers))
	targetUserIdMapping := make(map[string]*schema.User, 0)
	for _, targetAppUser := range targetAppUsers {
		targetAppUserIds = append(targetAppUserIds, targetAppUser.Identity)
		targetUserIdMapping[targetAppUser.Identity] = targetAppUser
	}
	for _, idpUser := range idpUsers {
		idpUserIdentity := n.GenerateIdpUserIdentity(idpUser.Identity)
		if !util.InArray(idpUserIdentity, targetAppUserIds) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"new_user":       idpUser,
			}).Debugf("Existence of new users")
			createUsers = append(createUsers, idpUser)
			continue
		}
		targetAppUser := targetUserIdMapping[idpUserIdentity]
		copyIdpUser := n.copyUser(idpUser, targetAppUser.RoleIds)
		if schema.UserIsChanged(targetAppUser, copyIdpUser) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"old_user":       targetAppUser,
				"new_user":       idpUser,
			}).Debugf("Existence of updated users")
			updateUsers = append(updateUsers, copyIdpUser)
		}
	}
	return
}

func (n *nexusApp) CompareGroups(idpGroups []*schema.Group, targetAppGroups []*schema.Group) (createGroups []*schema.Group, updateGroups []*schema.Group) {
	targetAppGroupIds := make([]string, 0, len(targetAppGroups))
	targetGroupIdMapping := make(map[string]*schema.Group, 0)
	for _, targetAppGroup := range targetAppGroups {
		targetAppGroupIds = append(targetAppGroupIds, targetAppGroup.Identity)
		targetGroupIdMapping[targetAppGroup.Identity] = targetAppGroup
	}
	for _, idpGroup := range idpGroups {
		idpGroupIdentity := n.GenerateIdpGroupIdentity(schema.NamespaceGroup, idpGroup.Identity)
		if !util.InArray(idpGroupIdentity, targetAppGroupIds) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"new_group":      idpGroup,
			}).Debugf("Existence of new group")
			createGroups = append(createGroups, idpGroup)
			continue
		}
		targetAppGroup := targetGroupIdMapping[idpGroupIdentity]
		newGroup := n.copyGroup(*idpGroup, *targetAppGroup)
		if schema.GroupIsChanged(targetAppGroup, newGroup) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"old_group":      targetAppGroup,
				"new_group":      idpGroup,
			}).Debugf("Existence of updated group")
			updateGroups = append(updateGroups, newGroup)
		}
	}
	return
}

func (n *nexusApp) CompareProjects(idpProjects []*schema.Project, targetAppProjects []*schema.Project) (createProjects []*schema.Project, updateProjects []*schema.Project) {
	targetAppProjectIds := make([]string, 0, len(targetAppProjects))
	targetProjectIdMapping := make(map[string]*schema.Project, 0)
	for _, targetAppProject := range targetAppProjects {
		targetAppProjectIds = append(targetAppProjectIds, targetAppProject.Identity)
		targetProjectIdMapping[targetAppProject.Identity] = targetAppProject
	}
	for _, idpProject := range idpProjects {
		idpProjectIdentity := n.GenerateIdpProjectIdentity(idpProject.Identity)
		if !util.InArray(idpProjectIdentity, targetAppProjectIds) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"new_project":    idpProject,
			}).Debugf("Existence of new project")
			createProjects = append(createProjects, idpProject)
			continue
		}
		targetAppProject := targetProjectIdMapping[idpProjectIdentity]
		if schema.ProjectIsChanged(targetAppProject, idpProject) {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": n.Kind(),
				"targetapp_name": n.GetName(),
				"old_project":    targetAppProject,
				"new_project":    idpProject,
			}).Debugf("Existence of updated project")
			updateProjects = append(updateProjects, idpProject)
		}
	}
	return
}

func (n *nexusApp) SyncGroupMember(ctx context.Context, idpGroupMembers []*schema.GroupMember, targetAppGroupMembers []*schema.GroupMember) error {
	idpUsers := schema.GroupMembersToUsers(idpGroupMembers)
	targetAppUsers := schema.GroupMembersToUsers(targetAppGroupMembers)
	idpUserIds := make(map[string][]string, 0)
	idpUserIdToOrgIdMapping := make(map[string]string)
	targetAppUserIds := make(map[string][]string, 0)
	targetAppOnlyGIds := make(map[string][]string, 0)
	for _, idpUser := range idpUsers {
		idpUserRoleIds := make([]string, 0)
		for _, roleId := range idpUser.RoleIds {
			idpUserRoleIds = append(idpUserRoleIds, n.GenerateIdpGroupIdentity(schema.NamespaceGroup, roleId))
		}
		identity := n.GenerateIdpUserIdentity(idpUser.Identity)
		idpUserIdToOrgIdMapping[identity] = idpUser.Identity
		idpUserIds[identity] = idpUserRoleIds
	}
	for _, targetAppUser := range targetAppUsers {
		targetAppUserRoleIds := make([]string, 0)
		onlyGIds := make([]string, 0)
		for _, roleId := range targetAppUser.RoleIds {
			knri := schema.StringToKNRI(roleId)
			if !knri.IsEmpty() && knri.Kind == n.idp.Kind().Tostring() && knri.Name == n.idp.GetName() && knri.RoleKind == schema.NamespaceGroup {
				onlyGIds = append(onlyGIds, roleId)
			}
			targetAppUserRoleIds = append(targetAppUserRoleIds, roleId)
		}
		targetAppUserIds[targetAppUser.Identity] = targetAppUserRoleIds
		targetAppOnlyGIds[targetAppUser.Identity] = onlyGIds
	}
	updateUsers := make([]*schema.User, 0)
	for identity, idpUserRoleIds := range idpUserIds {
		nexusRoleIds, ok := targetAppOnlyGIds[identity]
		if !ok {
			continue
		}
		sort.Strings(idpUserRoleIds)
		sort.Strings(nexusRoleIds)
		isUpdate := false
		if !cmp.Equal(idpUserRoleIds, nexusRoleIds) {
			isUpdate = true
		}
		if isUpdate {
			newRoleIds := make([]string, 0)
			addRoleIds := util.DiffArray(idpUserRoleIds, nexusRoleIds)
			delRoleIds := util.DiffArray(nexusRoleIds, idpUserRoleIds)
			delRoleIdMap := make(map[string]struct{})
			for _, roleId := range delRoleIds {
				delRoleIdMap[roleId] = struct{}{}
			}
			for _, roleId := range targetAppUserIds[identity] {
				if _, ok := delRoleIdMap[roleId]; !ok {
					newRoleIds = append(newRoleIds, roleId)
				}
			}
			newRoleIds = append(newRoleIds, addRoleIds...)
			// query new lastest user data
			user, err := n.idp.GetStaticUserById(idpUserIdToOrgIdMapping[identity])
			if err != nil {
				return err
			}
			// local dev, except `gitlab-gitlab1-group-48``
			if v := os.Getenv("EXCEPT_GROUP_MEMBER_ID"); len(v) > 0 {
				util.DeleteArrayItem(v, newRoleIds)
			}

			user.RoleIds = newRoleIds
			updateUsers = append(updateUsers, user)
		}
	}
	for _, updateUser := range updateUsers {
		err := n.UpdateUser(ctx, updateUser.Identity, updateUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *nexusApp) GroupBindingProjects(ctx context.Context, idpProjects []*schema.Project) error {
	groups, err := n.groupBindingProjectsHandle(ctx, idpProjects)
	if err != nil {
		return err
	}
	for _, group := range groups {
		role := n.idp2NexusConverter.IdpGroup2NexusRole(group.Identity, group)
		role.Roles = group.ChildIds
		err := n.client.Security.Role.Update(group.Identity, *role)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *nexusApp) WrappingUpAfterGroupSync(ctx context.Context) error {
	// Update father-son relationship
	idpGroups, err := n.idp.GetGroups(ctx)
	if err != nil {
		return err
	}
	nexusGroups, err := n.getLatestGroups()
	if err != nil {
		return err
	}
	nexusGroupMapping := make(map[string]*schema.Group)
	nexusGroupIdChildIds := make(map[string][]string)
	nexusGroupIdChildIdMapping := make(map[string]map[string]struct{})
	for _, nexusGroup := range nexusGroups {
		nexusGroupMapping[nexusGroup.Identity] = nexusGroup
		childIdMapping := make(map[string]struct{})
		for _, childId := range nexusGroup.ChildIds {
			childIdMapping[childId] = struct{}{}
		}
		nexusGroupIdChildIdMapping[nexusGroup.Identity] = childIdMapping
		nexusGroupIdChildIds[nexusGroup.Identity] = nexusGroup.ChildIds
	}

	idpGroupIdChildIds := make(map[string][]string)
	for _, idpGroup := range idpGroups {
		idpGroupIdentity := n.GenerateIdpGroupIdentity(idpGroup.Kind, idpGroup.Identity)
		childIds := make([]string, 0)
		for _, childId := range idpGroup.ChildIds {
			childIds = append(childIds, n.GenerateIdpGroupIdentity(idpGroup.Kind, childId))
		}
		idpGroupIdChildIds[idpGroupIdentity] = append(idpGroupIdChildIds[idpGroupIdentity], childIds...)
	}

	updateGroupIds := make([]string, 0)
	for identity, nexusChildIdMapping := range nexusGroupIdChildIdMapping {
		idpChildIds, ok := idpGroupIdChildIds[identity]
		if !ok {
			continue
		}
		if idpChildIds == nil {
			idpChildIds = make([]string, 0)
		}
		sort.Strings(idpChildIds)
		sort.Strings(nexusGroupMapping[identity].ChildIds)
		isChanged := false
		// ignore project role
		if !cmp.Equal(idpChildIds, nexusGroupMapping[identity].ChildIds, cmpopts.IgnoreSliceElements(func(item string) bool {
			if strings.Contains(item, "project") {
				return true
			}
			return false
		})) {
			isChanged = true
		}
		for _, childId := range idpChildIds {
			if _, ok := nexusChildIdMapping[childId]; !ok {
				nexusGroupMapping[identity].ChildIds = append(nexusGroupMapping[identity].ChildIds, childId)
			}
		}
		if isChanged {
			// intersect
			nexusGroupMapping[identity].ChildIds = util.Intersect(idpChildIds, nexusGroupMapping[identity].ChildIds)
			updateGroupIds = append(updateGroupIds, identity)
		}
	}

	for _, updateGroupId := range updateGroupIds {
		err := n.UpdateGroup(ctx, updateGroupId, nexusGroupMapping[updateGroupId])
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *nexusApp) getGroupIdProjectIdsRelationMapping(ctx context.Context) (map[string][]string, error) {
	err := n.newClient()
	if err != nil {
		return nil, err
	}
	list, err := n.client.Security.Role.List()
	if err != nil {
		return nil, err
	}
	groups := n.nexus2IdpConverter.ToNormalRoles(n.idp.Kind().Tostring(), n.idp.GetName(), list)
	result := make(map[string][]string)
	for _, group := range groups {
		for _, child := range group.ChildIds {
			if strings.Contains(child, schema.NamespaceProject) {
				result[group.Identity] = append(result[group.Identity], child)
			}
		}
	}
	return result, nil
}

func (n *nexusApp) groupBindingProjectsHandle(ctx context.Context, idpProjects []*schema.Project) ([]*schema.Group, error) {
	nexusGroups, err := n.getLatestGroups()
	if err != nil {
		return nil, err
	}
	result := make([]*schema.Group, 0)
	groupIdProjectIdMap := make(map[string][]string)
	for _, idpProject := range idpProjects {
		if devProject := os.Getenv("DEV_PROJECTNAME_PREFIX"); len(devProject) > 0 && !strings.Contains(idpProject.Name, devProject) {
			log.Loger.Debugf("debug project_name :%v", devProject)
			continue
		}
		groupId := n.GenerateIdpGroupIdentity(idpProject.Namespace.Kind, idpProject.Namespace.Identity)
		projectId := n.GenerateIdpProjectIdentity(idpProject.Identity)
		groupIdProjectIdMap[groupId] = append(groupIdProjectIdMap[groupId], projectId)
	}

	//nexusChildGroupsMapping := make(map[string][]string)

	// exist  group 、 namespace
	nexusGroupMapping := make(map[string][]string)
	for _, nexusGroup := range nexusGroups {
		for _, childId := range nexusGroup.ChildIds {
			nexusGroupMapping[nexusGroup.Identity] = append(nexusGroupMapping[nexusGroup.Identity], childId)
		}
	}

	idChildIdMap := make(map[string]map[string]struct{})
	idMap := make(map[string]*schema.Group)
	for _, nexusGroup := range nexusGroups {
		childIdMap := make(map[string]struct{})
		for _, childId := range nexusGroup.ChildIds {
			childIdMap[childId] = struct{}{}
		}
		idChildIdMap[nexusGroup.Identity] = childIdMap
		idMap[nexusGroup.Identity] = nexusGroup
	}
	for groupId, projectIds := range groupIdProjectIdMap {
		nexusGroup, ok := idMap[groupId]
		if !ok {
			continue
		}
		existNewChild := false
		for _, projectId := range projectIds {
			_, ok := idChildIdMap[groupId][projectId]
			if !ok {
				existNewChild = true
				nexusGroup.ChildIds = append(nexusGroup.ChildIds, projectId)
			}
		}
		if existNewChild {
			result = append(result, nexusGroup)
		}
	}
	return result, nil
}

func (n *nexusApp) copyUser(idpUser *schema.User, roleIds []string) *schema.User {
	u := &schema.User{
		BaseEntity: schema.BaseEntity{
			Identity:    idpUser.Identity,
			Name:        idpUser.Name,
			Description: idpUser.Description,
		},
		Username:    idpUser.Username,
		Email:       idpUser.Email,
		AvatarURL:   idpUser.AvatarURL,
		Mobile:      idpUser.Mobile,
		NamespaceId: idpUser.NamespaceId,
		RoleIds:     roleIds,
	}
	return u
}

func (n *nexusApp) copyGroup(idpGroup schema.Group, targetAppGroup schema.Group) *schema.Group {
	g := &schema.Group{
		BaseEntity: schema.BaseEntity{
			Identity:    targetAppGroup.Identity,
			Name:        idpGroup.Name,
			Description: idpGroup.Description,
		},
		Kind:     idpGroup.Kind,
		ChildIds: make([]string, 0),
	}
	idpGroupChildIds := make([]string, 0)
	for _, childId := range idpGroup.ChildIds {
		idpGroupChildIds = append(idpGroupChildIds, n.GenerateIdpGroupIdentity(schema.NamespaceGroup, childId))
	}
	projectRoleIds := make([]string, 0)
	for _, childId := range targetAppGroup.ChildIds {
		if strings.Contains(childId, schema.NamespaceProject) {
			projectRoleIds = append(projectRoleIds, childId)
		}
	}
	isChanged := false
	// ignore project role
	if !cmp.Equal(idpGroupChildIds, targetAppGroup.ChildIds, cmpopts.IgnoreSliceElements(func(item string) bool {
		if strings.Contains(item, "project") {
			return true
		}
		return false
	})) {
		isChanged = true
	}
	for _, childId := range idpGroupChildIds {
		if !util.InArray(childId, targetAppGroup.ChildIds) {
			targetAppGroup.ChildIds = append(targetAppGroup.ChildIds, childId)
		}
	}
	if isChanged {
		// intersect
		targetAppGroup.ChildIds = util.Intersect(idpGroupChildIds, targetAppGroup.ChildIds)
	}
	targetAppGroup.ChildIds = append(targetAppGroup.ChildIds, projectRoleIds...)
	if len(targetAppGroup.ChildIds) > 0 {
		g.ChildIds = targetAppGroup.ChildIds
	}
	return g
}
