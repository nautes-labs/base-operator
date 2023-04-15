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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nautes-labs/base-operator/pkg/idp"
	"github.com/nautes-labs/base-operator/pkg/log"
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/target"
	"github.com/nautes-labs/base-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// read idp data func signature
type readIdpDataHandleFuncSignature func() error

type readTargetAppDataHandleFuncSignature func(target.TargetApp) error

type SyncLogicService struct {
	ctx                          context.Context
	idp                          idp.Idp
	targetApps                   []target.TargetApp
	readIdpDataHandleFuncs       []readIdpDataHandleFuncSignature
	readTargetAppDataHandleFuncs []readTargetAppDataHandleFuncSignature
	//
	targetMapping     map[target.TargetAppKindName]target.TargetApp
	idpUsers          []*schema.User
	idpGroups         []*schema.Group
	idpProjects       []*schema.Project
	idpGroupMembers   []*schema.GroupMember
	idpProjectMembers []*schema.ProjectMember
	//
	targetAppUsersMapping       map[target.TargetAppKindName][]*schema.User
	targetAppGroupsMapping      map[target.TargetAppKindName][]*schema.Group
	targetAppProjectsMapping    map[target.TargetAppKindName][]*schema.Project
	targetAppGroupMemberMapping map[target.TargetAppKindName][]*schema.GroupMember
	//
	createTargetAppUsersMapping    map[target.TargetAppKindName][]*schema.User
	updateTargetAppUsersMapping    map[target.TargetAppKindName][]*schema.User
	createTargetAppGroupsMapping   map[target.TargetAppKindName][]*schema.Group
	updateTargetAppGroupsMapping   map[target.TargetAppKindName][]*schema.Group
	createTargetAppProjectsMapping map[target.TargetAppKindName][]*schema.Project
	updateTargetAppProjectsMapping map[target.TargetAppKindName][]*schema.Project
	//
	createTargetAppGroupMembersMapping   map[string][]*schema.GroupMember
	updateTargetAppGroupMembersMapping   map[string][]*schema.GroupMember
	createTargetAppProjectMembersMapping map[string][]*schema.ProjectMember
	updateTargetAppProjectMembersMapping map[string][]*schema.ProjectMember
	result                               *SyncLogicResult
}

// new service instance
func NewSyncLogicService(ctx context.Context) *SyncLogicService {
	result := &SyncLogicResult{
		Brief:  make([]*SyncLogicResultItem, 0),
		Detail: make(map[target.TargetAppKindName][]*SyncLogicResultItem, 0),
	}
	svc := &SyncLogicService{
		ctx:                            ctx,
		result:                         result,
		targetMapping:                  make(map[target.TargetAppKindName]target.TargetApp, 0),
		targetAppUsersMapping:          make(map[target.TargetAppKindName][]*schema.User, 0),
		createTargetAppUsersMapping:    make(map[target.TargetAppKindName][]*schema.User, 0),
		updateTargetAppUsersMapping:    make(map[target.TargetAppKindName][]*schema.User, 0),
		targetAppGroupsMapping:         make(map[target.TargetAppKindName][]*schema.Group, 0),
		createTargetAppGroupsMapping:   make(map[target.TargetAppKindName][]*schema.Group, 0),
		updateTargetAppGroupsMapping:   make(map[target.TargetAppKindName][]*schema.Group, 0),
		targetAppProjectsMapping:       make(map[target.TargetAppKindName][]*schema.Project, 0),
		createTargetAppProjectsMapping: make(map[target.TargetAppKindName][]*schema.Project, 0),
		updateTargetAppProjectsMapping: make(map[target.TargetAppKindName][]*schema.Project, 0),
		targetAppGroupMemberMapping:    make(map[target.TargetAppKindName][]*schema.GroupMember, 0),
	}
	svc.registerReadIdpDataHandleFunc(
		svc.readIdpUsers,
		svc.readIdpGroups,
		svc.readIdpProjects,
	)
	svc.registerReadTargetAppDataHandleFunc(
		svc.readTargetAppUsers,
		svc.readTargetAppGroups,
		svc.readTargetAppProjects,
	)
	return svc
}

// Service Entry Invoke Method
func (s *SyncLogicService) Run() error {
	defer util.PanicTrace()

	err := s.readData()
	if err != nil {
		return err
	}

	s.userDataHandle()
	s.groupDataHandle()
	s.projectDataHandle()

	err = s.writeTargetAppsData()
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLogicService) ClearTargetAppData() error {
	return nil
}

// register idp read method
func (s *SyncLogicService) registerReadIdpDataHandleFunc(funcs ...readIdpDataHandleFuncSignature) {
	s.readIdpDataHandleFuncs = append(s.readIdpDataHandleFuncs, funcs...)
}

// register targetApp read method
func (s *SyncLogicService) registerReadTargetAppDataHandleFunc(funcs ...readTargetAppDataHandleFuncSignature) {
	s.readTargetAppDataHandleFuncs = append(s.readTargetAppDataHandleFuncs, funcs...)
}

// Get Sync Result
func (s *SyncLogicService) GetResult() *SyncLogicResult {
	return s.result
}

// inject idp
func (s *SyncLogicService) InjectIdp(idp idp.Idp) *SyncLogicService {
	s.idp = idp
	return s
}

// inject target App
func (s *SyncLogicService) InjectTargetApps(targetApps ...target.TargetApp) *SyncLogicService {
	s.targetApps = append(s.targetApps, targetApps...)
	for _, targetApp := range targetApps {
		s.targetMapping[targetApp.IdentityKey()] = targetApp
	}
	return s
}

func (s *SyncLogicService) readData() error {
	log.Loger.Infof("Enter synchronous data reading phase")
	errGroup := new(errgroup.Group)
	errGroup.Go(s.readIdpData)
	errGroup.Go(s.readTargetAppsData)
	if err := errGroup.Wait(); err != nil {
		s.result.addBrief(NewReadResourceFailItem(err.Error()))
		return err
	}
	s.result.addBrief(NewReadResourceSuccessItem())
	return nil
}

func (s *SyncLogicService) readIdpData() error {
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Start idp data reading phase")
	// concurrent start
	errGroup := errgroup.Group{}
	for _, f := range s.readIdpDataHandleFuncs {
		errGroup.Go(f)
	}
	if err := errGroup.Wait(); err != nil {
		log.Loger.WithField("idp_kind", s.idp.Kind()).
			WithField("idp_name", s.idp.Kind()).Errorf("Idp data read fail, err:%v", err)
		return err
	}
	// concurrent end
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Idp user 、group 、project data read success")
	err := s.readIdpGroupMembers()
	if err != nil {
		log.Loger.WithField("idp_kind", s.idp.Kind()).
			WithField("idp_name", s.idp.Kind()).Errorf("Idp group member data read fail, err:%v", err)
		return err
	}
	return nil
}

func (s *SyncLogicService) readTargetAppsData() error {
	wg := sync.WaitGroup{}
	doErrChan := make(chan error)
	wg.Add(len(s.targetApps))
	for _, targetApp := range s.targetApps {
		go func(targetApp target.TargetApp) {
			defer wg.Done()
			if err := s.readTargetAppData(targetApp); err != nil {
				log.Loger.WithField("targetapp_kind", targetApp.Kind()).
					WithField("targetapp_name", targetApp.GetName()).
					Errorf("read targetapp data fail, err:%v", err)
				doErrChan <- err
				return
			}
			log.Loger.WithField("targetapp_kind", targetApp.Kind()).
				WithField("targetapp_name", targetApp.GetName()).
				Infof("read targetapp data success")
		}(targetApp)
	}
	go func() {
		defer close(doErrChan)
		wg.Wait()
	}()

	AggregateErr := (error)(nil)
	for errItem := range doErrChan {
		AggregateErr = multierror.Append(AggregateErr, errItem)
	}

	return AggregateErr
}

func (s *SyncLogicService) readTargetAppData(targetApp target.TargetApp) error {
	// concurrent start
	doErrChan := make(chan error)
	wg := sync.WaitGroup{}
	for _, f := range s.readTargetAppDataHandleFuncs {
		wg.Add(1)
		go func(targetApp target.TargetApp, f readTargetAppDataHandleFuncSignature) {
			defer wg.Done()
			if err := f(targetApp); err != nil {
				doErrChan <- err
				return
			}
		}(targetApp, f)
	}

	go func() {
		defer close(doErrChan)
		wg.Wait()
	}()

	AggregateErr := (error)(nil)
	for errItem := range doErrChan {
		AggregateErr = multierror.Append(AggregateErr, errItem)
	}
	if AggregateErr != nil {
		return AggregateErr
	}
	// concurrent end
	return nil
}

func (s *SyncLogicService) readIdpUsers() error {
	defer util.PanicTrace()
	users, err := s.idp.GetUsers(s.ctx)
	if err != nil {
		return fmt.Errorf("read users fail, err:%w", err)
	}
	s.idpUsers = users
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Idp user data read success")
	return nil
}

func (s *SyncLogicService) readIdpGroups() error {
	defer util.PanicTrace()
	groups, err := s.idp.GetGroups(s.ctx)
	if err != nil {
		return fmt.Errorf("read groups fail, err:%w", err)
	}
	s.idpGroups = groups
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Idp group data read success")
	return nil
}

func (s *SyncLogicService) readIdpProjects() error {
	defer util.PanicTrace()
	projects, err := s.idp.GetProjects(s.ctx)
	if err != nil {
		return fmt.Errorf("read project fail, err:%w", err)
	}
	s.idpProjects = projects
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Idp project data read success")
	return nil
}

func (s *SyncLogicService) readIdpGroupMembers() error {
	defer util.PanicTrace()
	groupMembers, err := s.idp.GetAllGroupMembers(s.ctx, s.idpGroups, s.idpUsers)
	if err != nil {
		return fmt.Errorf("read group members fail, err:%w", err)
	}
	s.idpGroupMembers = groupMembers
	log.Loger.WithField("idp_kind", s.idp.Kind()).
		WithField("idp_name", s.idp.Kind()).
		Infof("Idp group member data read success")
	return nil
}

func (s *SyncLogicService) readTargetAppUsers(targetApp target.TargetApp) error {
	defer util.PanicTrace()
	users, err := targetApp.GetUsers(s.ctx)
	if err != nil {
		errMsg := fmt.Sprintf("read target users fail, err:%v", err)
		s.result.addDetail(targetApp.IdentityKey(), NewSyncUserFailItem(errMsg))
		return errors.New(errMsg)
	}
	s.targetAppUsersMapping[targetApp.IdentityKey()] = users
	log.Loger.WithField("targetapp_kind", targetApp.Kind()).
		WithField("targetapp_name", targetApp.GetName()).
		Infof("read targetapp user data success")
	return nil
}

func (s *SyncLogicService) readTargetAppGroups(targetApp target.TargetApp) error {
	defer util.PanicTrace()
	groups, err := targetApp.GetGroups(s.ctx)
	if err != nil {
		errMsg := fmt.Sprintf("read target groups fail, err:%v", err)
		s.result.addDetail(targetApp.IdentityKey(), NewSyncGroupFailItem(errMsg))
		return errors.New(errMsg)
	}
	s.targetAppGroupsMapping[targetApp.IdentityKey()] = groups
	log.Loger.WithField("targetapp_kind", targetApp.Kind()).
		WithField("targetapp_name", targetApp.GetName()).
		Infof("read targetapp group data success")
	return nil
}

func (s *SyncLogicService) readTargetAppProjects(targetApp target.TargetApp) error {
	defer util.PanicTrace()
	projects, err := targetApp.GetProjects(s.ctx)
	if err != nil {
		errMsg := fmt.Sprintf("read target projects fail, err:%v", err)
		s.result.addDetail(targetApp.IdentityKey(), NewSyncProjectFailItem(errMsg))
		return errors.New(errMsg)
	}
	s.targetAppProjectsMapping[targetApp.IdentityKey()] = projects
	log.Loger.WithField("targetapp_kind", targetApp.Kind()).
		WithField("targetapp_name", targetApp.GetName()).
		Infof("read targetapp projects data success")
	return nil
}

func (s *SyncLogicService) readTargetAppGroupMembers(targetApp target.TargetApp) error {
	defer util.PanicTrace()
	groupMembers, err := targetApp.GetGroupMembers(s.ctx)
	if err != nil {
		errMsg := fmt.Sprintf("read target group members fail, err:%v", err)
		s.result.addDetail(targetApp.IdentityKey(), NewSyncGroupMemberFailItem(errMsg))
		return errors.New(errMsg)
	}
	s.targetAppGroupMemberMapping[targetApp.IdentityKey()] = groupMembers
	log.Loger.WithField("targetapp_kind", targetApp.Kind()).
		WithField("targetapp_name", targetApp.GetName()).
		Infof("read targetapp group members data success")
	return nil
}

func (s *SyncLogicService) userDataHandle() {
	for targetIdentity, targetUsers := range s.targetAppUsersMapping {
		createUsers, updateUsers := s.targetMapping[targetIdentity].CompareUsers(s.idpUsers, targetUsers)
		s.createTargetAppUsersMapping[targetIdentity] = createUsers
		s.updateTargetAppUsersMapping[targetIdentity] = updateUsers
	}
	return
}

func (s *SyncLogicService) groupDataHandle() {
	for targetIdentity, targetGroups := range s.targetAppGroupsMapping {
		createUsers, updateUsers := s.targetMapping[targetIdentity].CompareGroups(s.idpGroups, targetGroups)
		s.createTargetAppGroupsMapping[targetIdentity] = createUsers
		s.updateTargetAppGroupsMapping[targetIdentity] = updateUsers
	}
	return
}

func (s *SyncLogicService) projectDataHandle() {
	for targetIdentity, targetProjects := range s.targetAppProjectsMapping {
		createProjects, updateProjects := s.targetMapping[targetIdentity].CompareProjects(s.idpProjects, targetProjects)
		s.createTargetAppProjectsMapping[targetIdentity] = createProjects
		s.updateTargetAppProjectsMapping[targetIdentity] = updateProjects
	}
	return
}

func (s *SyncLogicService) syncGroupMember() error {
	AggregateErr := (error)(nil)
	for targetIdentity, targetGroupMembers := range s.targetAppGroupMemberMapping {
		err := s.targetMapping[targetIdentity].SyncGroupMember(s.ctx, s.idpGroupMembers, targetGroupMembers)
		AggregateErr = multierror.Append(AggregateErr, err)
	}
	return AggregateErr
}

func (s *SyncLogicService) writeTargetAppsData() error {
	doErrChan := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(s.targetApps))
	for _, targetApp := range s.targetApps {
		go func(targetApp target.TargetApp) {
			defer wg.Done()
			if err := s.writeTargetAppData(targetApp); err != nil {
				doErrChan <- err
				log.Loger.WithFields(logrus.Fields{
					"idp_kind":       s.idp.Kind(),
					"idp_name":       s.idp.GetName(),
					"targetapp_kind": targetApp.Kind(),
					"targetapp_name": targetApp.GetName(),
				}).Errorf("idp data to targetapp fail, err:%v", err)
				return
			}
			log.Loger.WithFields(logrus.Fields{
				"idp_kind":       s.idp.Kind(),
				"idp_name":       s.idp.GetName(),
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Info("idp data to targetapp success")
		}(targetApp)
	}
	go func() {
		defer close(doErrChan)
		wg.Wait()
	}()

	AggregateErr := (error)(nil)
	for errItem := range doErrChan {
		AggregateErr = multierror.Append(AggregateErr, errItem)
	}

	return AggregateErr
}

func (s *SyncLogicService) writeTargetAppData(targetApp target.TargetApp) error {
	defer util.PanicTrace()
	err := (error)(nil)
	err = s.syncUser(targetApp)
	if err != nil {
		return err
	}
	err = s.syncGroup(targetApp)
	if err != nil {
		return err
	}
	err = s.syncProjects(targetApp)
	if err != nil {
		return err
	}
	// Get the latest target application group members
	err = s.readTargetAppGroupMembers(targetApp)
	if err != nil {
		return err
	}
	err = s.syncGroupMember()
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLogicService) syncUser(targetApp target.TargetApp) error {
	err := s.syncCreateUser(targetApp)
	if err != nil {
		return err
	}
	err = s.syncUpdateUser(targetApp)
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLogicService) syncCreateUser(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	createUsers := s.createTargetAppUsersMapping[targetIdentity]
	for _, createUser := range createUsers {
		if devUser := os.Getenv("DEV_USERNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(createUser.Name, devUser) {
			log.Loger.Debugf("debug user_name :%v", devUser)
			continue
		}
		err := targetApp.CreateUser(s.ctx, createUser)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("create user fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncUserFailItem(err.Error()))
			return err
		}
	}
	return nil
}

func (s *SyncLogicService) syncUpdateUser(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	updateUsers := s.updateTargetAppUsersMapping[targetIdentity]
	for _, updateUser := range updateUsers {
		if devUser := os.Getenv("DEV_USERNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(updateUser.Name, devUser) {
			log.Loger.Debugf("debug user_name :%v", devUser)
			continue
		}
		err := targetApp.UpdateUser(s.ctx, updateUser.Identity, updateUser)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("update user fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncUserFailItem(err.Error()))
			return err
		}
	}
	return nil
}

func (s *SyncLogicService) syncGroup(targetApp target.TargetApp) error {
	err := s.syncCreateGroup(targetApp)
	if err != nil {
		return err
	}
	err = s.syncUpdateGroup(targetApp)
	if err != nil {
		return err
	}
	err = targetApp.WrappingUpAfterGroupSync(s.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLogicService) syncCreateGroup(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	createGroups := s.createTargetAppGroupsMapping[targetIdentity]
	for _, createGroup := range createGroups {
		if devUser := os.Getenv("DEV_GROUPNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(createGroup.Name, devUser) {
			log.Loger.Debugf("debug group_name :%v", devUser)
			continue
		}
		err := targetApp.CreateGroup(s.ctx, createGroup)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("create group fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncGroupFailItem(err.Error()))
			return err
		}
	}
	return nil
}

func (s *SyncLogicService) syncUpdateGroup(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	updateGroups := s.updateTargetAppGroupsMapping[targetIdentity]
	for _, updateGroup := range updateGroups {
		if devUser := os.Getenv("DEV_GROUPNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(updateGroup.Name, devUser) {
			log.Loger.Debugf("debug group_name :%v", devUser)
			continue
		}
		err := targetApp.UpdateGroup(s.ctx, updateGroup.Identity, updateGroup)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("update group fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncGroupFailItem(err.Error()))
			return err
		}
	}
	return nil
}

func (s *SyncLogicService) syncProjects(targetApp target.TargetApp) error {
	err := (error)(nil)
	err = s.syncCreateProject(targetApp)
	if err != nil {
		return err
	}
	err = targetApp.GroupBindingProjects(s.ctx, s.idpProjects)
	if err != nil {
		return err
	}
	err = s.syncUpdateProject(targetApp)
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLogicService) syncCreateProject(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	createProjects := s.createTargetAppProjectsMapping[targetIdentity]
	for _, createProject := range createProjects {
		if devUser := os.Getenv("DEV_PROJECTNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(createProject.Name, devUser) {
			log.Loger.Debugf("debug project_name :%v", devUser)
			continue
		}
		err := targetApp.CreateProject(s.ctx, createProject)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("create project fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncProjectFailItem(err.Error()))
			return err
		}
	}
	return nil
}

func (s *SyncLogicService) syncUpdateProject(targetApp target.TargetApp) error {
	targetIdentity := targetApp.IdentityKey()
	updateProjects := s.updateTargetAppProjectsMapping[targetIdentity]
	for _, updateProject := range updateProjects {
		if devUser := os.Getenv("DEV_PROJECTNAME_PREFIX"); len(devUser) > 0 && !strings.Contains(updateProject.Name, devUser) {
			log.Loger.Debugf("debug project_name :%v", devUser)
			continue
		}
		err := targetApp.UpdateProject(s.ctx, updateProject.Identity, updateProject)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"targetapp_kind": targetApp.Kind(),
				"targetapp_name": targetApp.GetName(),
			}).Errorf("update project fail, err:%v", err)
			s.result.addDetail(targetIdentity, NewSyncProjectFailItem(err.Error()))
			return err
		}
	}
	return nil
}
