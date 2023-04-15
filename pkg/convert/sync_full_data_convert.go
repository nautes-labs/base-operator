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

package convert

//type Gitlab2Idp struct {
//	idpUsers          []*schema.User
//	idpGroups         []*schema.Group
//	idpProjects       []*schema.Project
//	idpGroupMembers   []*schema.GroupMember
//	idpProjectMembers []*schema.ProjectMember
//}
//
//func NewGitlab2Idp(idpUsers []*schema.User, idpGroups []*schema.Group, idpProjects []*schema.Project, idpGroupMembers []*schema.GroupMember, idpProjectMembers []*schema.ProjectMember) *Gitlab2Idp {
//	return &Gitlab2Idp{idpUsers: idpUsers, idpGroups: idpGroups, idpProjects: idpProjects, idpGroupMembers: idpGroupMembers, idpProjectMembers: idpProjectMembers}
//}
//
//type Nexus2Idp struct {
//	idpUsers          []*schema.User
//	idpGroups         []*schema.Group
//	idpProjects       []*schema.Project
//	idpGroupMembers   []*schema.GroupMember
//	idpProjectMembers []*schema.ProjectMember
//}
//
//func NewNexus2Idp(idpUsers []*schema.User, idpGroups []*schema.Group, idpProjects []*schema.Project, idpGroupMembers []*schema.GroupMember, idpProjectMembers []*schema.ProjectMember) *Nexus2Idp {
//	return &Nexus2Idp{idpUsers: idpUsers, idpGroups: idpGroups, idpProjects: idpProjects, idpGroupMembers: idpGroupMembers, idpProjectMembers: idpProjectMembers}
//}
//
//func getIdpDataFromGitlab() *Gitlab2Idp {
//	idpUsers := make([]*schema.User, 0)
//	idpGroups := make([]*schema.Group, 0)
//	idpProjects := make([]*schema.Project, 0)
//	idpGroupMembers := make([]*schema.GroupMember, 0)
//	idpProjectMembers := make([]*schema.ProjectMember, 0)
//	gitlabUsers := []*gitlab.User{
//		{ID: 1, Username: "u1", Email: "101@qq.com", Name: "z1", AvatarURL: "xxxxxxxxx111"},
//		{ID: 2, Username: "u2", Email: "102@qq.com", Name: "z2", AvatarURL: "xxxxxxxxx112"},
//		{ID: 3, Username: "u3", Email: "103@qq.com", Name: "z3", AvatarURL: "xxxxxxxxx113"},
//		{ID: 4, Username: "u4", Email: "104@qq.com", Name: "z4", AvatarURL: "xxxxxxxxx114"},
//		{ID: 5, Username: "u5", Email: "105@qq.com", Name: "z5", AvatarURL: "xxxxxxxxx115"},
//		{ID: 6, Username: "u6", Email: "106@qq.com", Name: "z6", AvatarURL: "xxxxxxxxx116"},
//		{ID: 7, Username: "u7", Email: "107@qq.com", Name: "z7", AvatarURL: "xxxxxxxxx117"},
//	}
//	gitlabGroups := []*gitlab.Group{
//		{ID: 1, Name: "g1", Description: "desc", ParentID: 0},
//		{ID: 2, Name: "g2", Description: "desc", ParentID: 1},
//		{ID: 3, Name: "g3", Description: "desc", ParentID: 2},
//		{ID: 4, Name: "g4", Description: "desc", ParentID: 0},
//	}
//	gitlabProjects := []*gitlab.Project{
//		{ID: 1, Name: "p1", Description: "desc", Namespace: &gitlab.ProjectNamespace{ID: 1, Kind: "group", ParentID: 0}},
//		{ID: 2, Name: "p2", Description: "desc", Namespace: &gitlab.ProjectNamespace{ID: 2, Kind: "group", ParentID: 1}},
//		{ID: 3, Name: "p3", Description: "desc", Namespace: &gitlab.ProjectNamespace{ID: 3, Kind: "group", ParentID: 2}},
//		{ID: 4, Name: "p4", Description: "desc", Namespace: &gitlab.ProjectNamespace{ID: 3, Kind: "group", ParentID: 2}},
//		{ID: 5, Name: "p5", Description: "desc", Namespace: &gitlab.ProjectNamespace{ID: 100, Kind: "user", ParentID: 0}},
//	}
//	gitlabGroupMembersMapping := map[int][]*gitlab.GroupMember{
//		1: {{ID: 1}, {ID: 2}, {ID: 3}},
//		2: {{ID: 2}, {ID: 3}},
//		3: {{ID: 5}, {ID: 7}},
//		4: {{ID: 6}},
//	}
//	gitlabProjectMembersMapping := map[int][]*gitlab.ProjectMember{
//		1: {{ID: 5}},
//		2: {{ID: 6}},
//		3: {{ID: 7}},
//		4: {{ID: 7}},
//		5: {{ID: 7}},
//	}
//	for _, user := range gitlabUsers {
//		idpUser := gitlab2.GitlabUser2IdpUser(user)
//		idpUsers = append(idpUsers, idpUser)
//	}
//	for _, group := range gitlabGroups {
//		idpGroup := gitlab2.GitlabGroup2IdpGroup(group)
//		idpGroups = append(idpGroups, idpGroup)
//	}
//	for _, project := range gitlabProjects {
//		idpProject := gitlab2.GitlabProject2IdpProject(project)
//		idpProjects = append(idpProjects, idpProject)
//	}
//	for groupId, groupMembers := range gitlabGroupMembersMapping {
//		for _, groupMember := range groupMembers {
//			idpGroupMember := gitlab2.GitlabGroupMember2IdpGroupMember(cast.ToString(groupId), groupMember)
//			idpGroupMembers = append(idpGroupMembers, idpGroupMember)
//		}
//	}
//	for projectId, projectMembers := range gitlabProjectMembersMapping {
//		for _, projectMember := range projectMembers {
//			idpProjectMember := gitlab2.GitlabProjectMember2IdpProjectMember(cast.ToString(projectId), projectMember)
//			idpProjectMembers = append(idpProjectMembers, idpProjectMember)
//		}
//	}
//	return NewGitlab2Idp(idpUsers, idpGroups, idpProjects, idpGroupMembers, idpProjectMembers)
//}
//
//func getIdpDataFromNexus() *Nexus2Idp {
//	idpUsers := make([]*schema.User, 0)
//	idpGroups := make([]*schema.Group, 0)
//	idpProjects := make([]*schema.Project, 0)
//	idpGroupMembers := make([]*schema.GroupMember, 0)
//	idpProjectMembers := make([]*schema.ProjectMember, 0)
//	nexusUsers := []*nexus.User{
//		{UserID: "1", LastName: "z1", FirstName: "z2", EmailAddress: "999@qq.com", Roles: []string{"group-1"}},
//		{UserID: "2", LastName: "z100", FirstName: "z200", EmailAddress: "771@qq.com", Roles: []string{"project-2"}},
//	}
//	nexusRoles := []*nexus.Role{
//		{ID: "group-1", Name: "r1", Description: "desc", Roles: []string{"group-2", "group-3"}},
//		{ID: "group-2", Name: "r2", Description: "desc", Roles: []string{"project-1"}},
//		{ID: "group-3", Name: "r3", Description: "desc", Roles: []string{"project-2"}},
//		{ID: "project-1", Name: "r4", Description: "desc", Roles: []string{}},
//		{ID: "project-2", Name: "r5", Description: "desc", Roles: []string{}},
//	}
//	for _, user := range nexusUsers {
//		idpUsers = append(idpUsers, NexusUser2IdpUser(user))
//		idpGroupMembers = append(idpGroupMembers, NexusUser2IdpGroupMember(user)...)
//		idpProjectMembers = append(idpProjectMembers, NexusUser2IdpProjectMember(user)...)
//	}
//	idpGroups = append(idpGroups, NexusRoles2IdpGroup(nexusRoles)...)
//	idpProjects = append(idpProjects, NexusRoles2IdpProject(nexusRoles)...)
//	return NewNexus2Idp(idpUsers, idpGroups, idpProjects, idpGroupMembers, idpProjectMembers)
//}
//
//func firstSync() {
//	_ = getIdpDataFromGitlab()
//	//TODO idp to nexus data
//	//TODO nexus data write
//}
//
//func Full() {
//	nexusData := getIdpDataFromNexus()
//	gitlabData := getIdpDataFromGitlab()
//	nexusUserIdMapping := make(map[string]*schema.User)
//	nexusGroupIdMapping := make(map[string]*schema.Group)
//	for _, u := range nexusData.idpUsers {
//		nexusUserIdMapping[u.Id] = u
//	}
//	for _, g := range nexusData.idpGroups {
//		nexusGroupIdMapping[g.Id] = g
//	}
//	needCreateUsers := make([]*schema.User, 0)
//	needModifyUsers := make([]*schema.User, 0)
//	needCreateGroups := make([]*schema.Group, 0)
//	needModifyGroups := make([]*schema.Group, 0)
//	for _, u1 := range gitlabData.idpUsers {
//		item, ok := nexusUserIdMapping[u1.Id]
//		if !ok {
//			needCreateUsers = append(needCreateUsers, item)
//		} else {
//			needModifyUsers = append(needModifyUsers, item)
//		}
//	}
//
//	for _, g1 := range gitlabData.idpGroups {
//		item, ok := nexusGroupIdMapping[g1.Id]
//		if !ok {
//			needCreateGroups = append(needCreateGroups, item)
//		} else {
//			needModifyGroups = append(needModifyGroups, item)
//		}
//	}
//	fmt.Println("111111111")
//}
