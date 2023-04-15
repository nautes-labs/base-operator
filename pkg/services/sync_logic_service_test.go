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

import (
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/nautes-labs/base-operator/pkg/idp"
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
	"github.com/nautes-labs/base-operator/pkg/target"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sync base data", func() {
	var (
		err                   error
		svc                   *SyncLogicService
		idpName               string
		idpApiServerUrl       string
		secretProvider        *secret_provider.SecretProvider
		idpMock               *idp.MockIdp
		targetAppname         string
		targetAppApiServerUrl string
		targetAppMock         *target.MockTargetApp
		idpUsers              []*schema.User
		idpGroups             []*schema.Group
		idpProjects           []*schema.Project
		targetUsers           []*schema.User
		targetGroups          []*schema.Group
		targetProjects        []*schema.Project
		//targetGroupMembers    []*schema.GroupMember
		//targetProjectMembers  []*schema.ProjectMember
	)
	BeforeEach(func() {
		err = nil
		svc = NewSyncLogicService(ctx)
		// idp
		idpName = "gitlab1"
		idpApiServerUrl = "https://github.com/api/v4"
		secretProvider = nil
		idpMock = idp.NewMockIdp(ctl)
		idpMock.EXPECT().Kind().Return(idp.GitlabIdpKind).AnyTimes()
		idpMock.EXPECT().SetName(idpName).AnyTimes()
		idpMock.EXPECT().GetName().AnyTimes()
		idpMock.EXPECT().SetApiServerUrl(idpApiServerUrl).AnyTimes()
		idpMock.EXPECT().SetSecretProvider(secretProvider).AnyTimes()
		svc.InjectIdp(idpMock)
		// targetapp
		targetAppname = "nexus1"
		targetAppApiServerUrl = "http://nexus.bluzin.io:8081"
		targetAppMock = target.NewMockTargetApp(ctl)
		targetAppMock.EXPECT().Kind().Return(target.NexusAppKind).AnyTimes()
		targetAppMock.EXPECT().SetName(targetAppname).AnyTimes()
		targetAppMock.EXPECT().GetName().AnyTimes()
		targetAppMock.EXPECT().IdentityKey().Return(target.TargetAppKindName{
			Kind: string(target.NexusAppKind),
			Name: targetAppname,
		}).AnyTimes()
		targetAppMock.EXPECT().SetApiServerUrl(targetAppApiServerUrl).AnyTimes()
		targetAppMock.EXPECT().SetSecretProvider(secretProvider).AnyTimes()
		targetAppMock.EXPECT().SetIdp(idpMock).AnyTimes()
		svc.InjectTargetApps(targetAppMock)
		idpUsers = []*schema.User{
			{BaseEntity: schema.BaseEntity{Identity: "100"}},
		}
		targetUsers = []*schema.User{
			{BaseEntity: schema.BaseEntity{Identity: "200"}},
		}
		idpGroups = []*schema.Group{
			{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
		}
		targetGroups = []*schema.Group{
			{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
		}
		idpProjects = []*schema.Project{
			{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi", Description: "x6666"}},
		}
		targetProjects = []*schema.Project{
			{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi", Description: "x3333"}},
		}
	})
	Context("Users", func() {
		It("Failed to get targetapp user", func() {
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp user", func() {
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpUsers()
			Expect(err).Should(HaveOccurred())
		})
		It("Get idp user successfully, failed to get targetapp user", func() {
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(idpUsers, nil)
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpUsers()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp user, get targetapp user successfully", func() {
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(nil, errors.New("timeout"))
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(targetUsers, nil)
			err = svc.readIdpUsers()
			Expect(err).Should(HaveOccurred())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to creating new user", func() {
			idpUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createUsers := []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(idpUsers, nil)
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(targetUsers, nil)
			err = svc.readIdpUsers()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareUsers(gomock.Any(), gomock.Any()).Return(createUsers, nil)
			svc.userDataHandle()
			targetAppMock.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncCreateUser(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Creating new user successfully", func() {
			idpUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createUsers := []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(idpUsers, nil)
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(targetUsers, nil)
			err = svc.readIdpUsers()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareUsers(gomock.Any(), gomock.Any()).Return(createUsers, nil)
			svc.userDataHandle()
			targetAppMock.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncCreateUser(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to update user", func() {
			idpUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateUsers := []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(idpUsers, nil)
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(targetUsers, nil)
			err = svc.readIdpUsers()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareUsers(gomock.Any(), gomock.Any()).Return(nil, updateUsers)
			svc.userDataHandle()
			targetAppMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncUpdateUser(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Update user successfully", func() {
			idpUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetUsers = []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateUsers := []*schema.User{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetUsers(gomock.Any()).Return(idpUsers, nil)
			targetAppMock.EXPECT().GetUsers(gomock.Any()).Return(targetUsers, nil)
			err = svc.readIdpUsers()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppUsers(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareUsers(gomock.Any(), gomock.Any()).Return(nil, updateUsers)
			svc.userDataHandle()
			targetAppMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncUpdateUser(targetAppMock)
			Expect(err).Should(BeNil())
		})
	})
	Context("Group", func() {
		It("Failed to get targetapp group", func() {
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp group", func() {
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpGroups()
			Expect(err).Should(HaveOccurred())
		})
		It("Get idp group successfully, failed to get targetapp group", func() {
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(idpGroups, nil)
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpGroups()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp group, get targetapp group successfully", func() {
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(nil, errors.New("timeout"))
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(targetGroups, nil)
			err = svc.readIdpGroups()
			Expect(err).Should(HaveOccurred())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to creating new group", func() {
			idpGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createGroups := []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(idpGroups, nil)
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(targetGroups, nil)
			err = svc.readIdpGroups()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareGroups(gomock.Any(), gomock.Any()).Return(createGroups, nil)
			svc.groupDataHandle()
			targetAppMock.EXPECT().CreateGroup(gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncCreateGroup(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Creating new group successfully", func() {
			idpGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createGroups := []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(idpGroups, nil)
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(targetGroups, nil)
			err = svc.readIdpGroups()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareGroups(gomock.Any(), gomock.Any()).Return(createGroups, nil)
			svc.groupDataHandle()
			targetAppMock.EXPECT().CreateGroup(gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncCreateGroup(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to update group", func() {
			idpGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateGroups := []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(idpGroups, nil)
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(targetGroups, nil)
			err = svc.readIdpGroups()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareGroups(gomock.Any(), gomock.Any()).Return(nil, updateGroups)
			svc.groupDataHandle()
			targetAppMock.EXPECT().UpdateGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncUpdateGroup(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Update group successfully", func() {
			idpGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetGroups = []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateGroups := []*schema.Group{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetGroups(gomock.Any()).Return(idpGroups, nil)
			targetAppMock.EXPECT().GetGroups(gomock.Any()).Return(targetGroups, nil)
			err = svc.readIdpGroups()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppGroups(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareGroups(gomock.Any(), gomock.Any()).Return(nil, updateGroups)
			svc.groupDataHandle()
			targetAppMock.EXPECT().UpdateGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncUpdateGroup(targetAppMock)
			Expect(err).Should(BeNil())
		})
	})
	Context("Project", func() {
		It("Failed to get targetapp project", func() {
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp project", func() {
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpProjects()
			Expect(err).Should(HaveOccurred())
		})
		It("Get idp project successfully, failed to get targetapp project", func() {
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(idpProjects, nil)
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(nil, errors.New("timeout"))
			err = svc.readIdpProjects()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Failed to get idp project, get targetapp project successfully", func() {
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(nil, errors.New("timeout"))
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(targetProjects, nil)
			err = svc.readIdpProjects()
			Expect(err).Should(HaveOccurred())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to creating new project", func() {
			idpProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createProjects := []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(idpProjects, nil)
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(targetProjects, nil)
			err = svc.readIdpProjects()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareProjects(gomock.Any(), gomock.Any()).Return(createProjects, nil)
			svc.projectDataHandle()
			targetAppMock.EXPECT().CreateProject(gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncCreateProject(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Creating new project successfully", func() {
			idpProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			targetProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "200", Name: "lisi"}},
			}
			createProjects := []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh"}},
			}
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(idpProjects, nil)
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(targetProjects, nil)
			err = svc.readIdpProjects()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareProjects(gomock.Any(), gomock.Any()).Return(createProjects, nil)
			svc.projectDataHandle()
			targetAppMock.EXPECT().CreateProject(gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncCreateProject(targetAppMock)
			Expect(err).Should(BeNil())
		})
		It("Failed to update project", func() {
			idpProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateProjects := []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(idpProjects, nil)
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(targetProjects, nil)
			err = svc.readIdpProjects()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareProjects(gomock.Any(), gomock.Any()).Return(nil, updateProjects)
			svc.projectDataHandle()
			targetAppMock.EXPECT().UpdateProject(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
			err = svc.syncUpdateProject(targetAppMock)
			Expect(err).Should(HaveOccurred())
		})
		It("Update project successfully", func() {
			idpProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			targetProjects = []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x3333"}},
			}
			updateProjects := []*schema.Project{
				{BaseEntity: schema.BaseEntity{Identity: "100", Name: "zhangxh", Description: "x6666"}},
			}
			idpMock.EXPECT().GetProjects(gomock.Any()).Return(idpProjects, nil)
			targetAppMock.EXPECT().GetProjects(gomock.Any()).Return(targetProjects, nil)
			err = svc.readIdpProjects()
			Expect(err).Should(BeNil())
			err = svc.readTargetAppProjects(targetAppMock)
			Expect(err).Should(BeNil())
			targetAppMock.EXPECT().CompareProjects(gomock.Any(), gomock.Any()).Return(nil, updateProjects)
			svc.projectDataHandle()
			targetAppMock.EXPECT().UpdateProject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err = svc.syncUpdateProject(targetAppMock)
			Expect(err).Should(BeNil())
		})
	})
	// Context("wrapping Up After Project", func() {
	// 	It("Failed", func() {
	// 		targetAppMock.EXPECT().GroupBindingProjects(gomock.Any(), gomock.Any()).Return(errors.New("timeout"))
	// 		err := svc.wrappingUpAfterSyncProject(targetAppMock)
	// 		Expect(err).Should(HaveOccurred())
	// 	})
	// 	It("Successfully", func() {
	// 		targetAppMock.EXPECT().GroupBindingProjects(gomock.Any(), gomock.Any()).Return(nil)
	// 		err := svc.wrappingUpAfterSyncProject(targetAppMock)
	// 		Expect(err).Should(BeNil())
	// 	})
	// })
})
