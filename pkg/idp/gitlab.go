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
	"fmt"
	"math"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nautes-labs/base-operator/pkg/convert/convert2idp"
	"github.com/nautes-labs/base-operator/pkg/schema"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
	"github.com/spf13/cast"
	"github.com/xanzy/go-gitlab"
)

const (
	gitlabUserPageSize        = 20
	gitlabGroupPageSize       = 20
	gitlabProjectPageSize     = 20
	gitlabGroupMemberPageSize = 20
)

var _ Idp = (*gitlabIdp)(nil)

type gitlabIdp struct {
	name           string
	apiServerUrl   string
	secretProvider *secret_provider.SecretProvider
	client         *gitlab.Client
	converter      *convert2idp.Gitlab2IdpConverter
	users          []*schema.User
	groups         []*schema.Group
	projects       []*schema.Project
}

func (g *gitlabIdp) Kind() IdpKind {
	return GitlabIdpKind
}

func (g *gitlabIdp) SetName(name string) {
	g.name = name
	return
}

func (g *gitlabIdp) GetName() string {
	return g.name
}

func (g *gitlabIdp) SetApiServerUrl(url string) {
	g.apiServerUrl = url
	return
}

func (g *gitlabIdp) SetSecretProvider(provider *secret_provider.SecretProvider) {
	g.secretProvider = provider
	return
}

func (g *gitlabIdp) GetUsers(ctx context.Context) ([]*schema.User, error) {
	err := g.newClient()
	if err != nil {
		return nil, fmt.Errorf("init gitlab client fail, err:【%w】", err)
	}
	page := 1
	opts := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabUserPageSize},
	}
	list, rsp, err := g.client.Users.ListUsers(opts)
	if err != nil {
		return nil, err
	}
	totalCount := rsp.TotalItems
	gitlabUsers := make([]*gitlab.User, 0)
	for _, user := range list {
		gitlabUsers = append(gitlabUsers, user)
	}
	totalPage := cast.ToInt(math.Ceil(cast.ToFloat64(totalCount) / cast.ToFloat64(gitlabUserPageSize)))
	doChan := make(chan interface{}, 1)
	wg := sync.WaitGroup{}
	for page = 2; page <= totalPage; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			opts := &gitlab.ListUsersOptions{
				ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabUserPageSize},
			}
			list, _, err := g.client.Users.ListUsers(opts)
			if err != nil {
				doChan <- err
				return
			}
			doChan <- list
			return
		}(page)
	}
	go func() {
		defer close(doChan)
		wg.Wait()
	}()
	for item := range doChan {
		switch assertValue := item.(type) {
		case error:
			return nil, assertValue
		case []*gitlab.User:
			gitlabUsers = append(gitlabUsers, assertValue...)
		}
	}
	result := make([]*schema.User, 0, len(gitlabUsers))
	for _, gitlabUser := range gitlabUsers {
		item := g.converter.ToIdpUser(gitlabUser)
		result = append(result, item)
	}
	g.users = result
	return result, nil
}

func (g *gitlabIdp) GetStaticUserById(id string) (*schema.User, error) {
	for _, user := range g.users {
		if user.Identity == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found, id:%s", id)
}

func (g *gitlabIdp) GetGroups(ctx context.Context) ([]*schema.Group, error) {
	if len(g.groups) > 0 {
		return g.groups, nil
	}
	err := g.newClient()
	if err != nil {
		return nil, fmt.Errorf("init gitlab client fail, err:【%w】", err)
	}
	page := 1
	opts := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabGroupPageSize},
	}
	list, rsp, err := g.client.Groups.ListGroups(opts)
	if err != nil {
		return nil, err
	}
	totalCount := rsp.TotalItems
	gitlabGroups := make([]*gitlab.Group, 0)
	for _, user := range list {
		gitlabGroups = append(gitlabGroups, user)
	}
	totalPage := cast.ToInt(math.Ceil(cast.ToFloat64(totalCount) / cast.ToFloat64(gitlabGroupPageSize)))
	doChan := make(chan interface{}, 1)
	wg := sync.WaitGroup{}
	for page = 2; page <= totalPage; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			opts := &gitlab.ListGroupsOptions{
				ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabGroupPageSize},
			}
			list, _, err := g.client.Groups.ListGroups(opts)
			if err != nil {
				doChan <- err
				return
			}
			doChan <- list
			return
		}(page)
	}
	go func() {
		defer close(doChan)
		wg.Wait()
	}()
	for item := range doChan {
		switch assertValue := item.(type) {
		case error:
			return nil, assertValue
		case []*gitlab.Group:
			gitlabGroups = append(gitlabGroups, assertValue...)
		}
	}
	groups := make([]*schema.Group, 0, len(gitlabGroups))
	for _, gitlabGroup := range gitlabGroups {
		item := g.converter.ToIdpGroup(gitlabGroup, schema.NamespaceGroup)
		groups = append(groups, item)
	}
	schema.RenderChildGroupIds(groups)
	g.groups = groups
	return groups, nil
}

func (g *gitlabIdp) GetProjects(ctx context.Context) ([]*schema.Project, error) {
	err := g.newClient()
	if err != nil {
		return nil, fmt.Errorf("init gitlab client fail, err:【%w】", err)
	}
	page := 1
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabProjectPageSize},
	}
	list, rsp, err := g.client.Projects.ListProjects(opts)
	if err != nil {
		return nil, err
	}
	totalCount := rsp.TotalItems
	gitlabProjects := make([]*gitlab.Project, 0)
	for _, user := range list {
		gitlabProjects = append(gitlabProjects, user)
	}
	totalPage := cast.ToInt(math.Ceil(cast.ToFloat64(totalCount) / cast.ToFloat64(gitlabProjectPageSize)))
	doChan := make(chan interface{}, 1)
	wg := sync.WaitGroup{}
	for page = 2; page <= totalPage; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			opts := &gitlab.ListProjectsOptions{
				ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabProjectPageSize},
			}
			list, _, err := g.client.Projects.ListProjects(opts)
			if err != nil {
				doChan <- err
				return
			}
			doChan <- list
			return
		}(page)
	}
	go func() {
		defer close(doChan)
		wg.Wait()
	}()
	for item := range doChan {
		switch assertValue := item.(type) {
		case error:
			return nil, assertValue
		case []*gitlab.Project:
			gitlabProjects = append(gitlabProjects, assertValue...)
		}
	}
	projects := make([]*schema.Project, 0, len(gitlabProjects))
	for _, gitlabProject := range gitlabProjects {
		item := g.converter.ToIdpProject(gitlabProject)
		projects = append(projects, item)
	}
	g.projects = projects
	return projects, nil
}

func (g *gitlabIdp) GetStaticProjects() []*schema.Project {
	return g.projects
}

func (g *gitlabIdp) GetAllGroupMembers(ctx context.Context, groups []*schema.Group, users []*schema.User) ([]*schema.GroupMember, error) {
	result := make([]*schema.GroupMember, 0)
	wg := sync.WaitGroup{}
	doChan := make(chan interface{})
	for _, group := range groups {
		wg.Add(1)
		go func(ctx context.Context, group *schema.Group, user *schema.User) {
			defer wg.Done()
			groupMembers, err := g.GetGroupMembers(ctx, group, nil)
			if err != nil {
				doChan <- err
				return
			}
			doChan <- groupMembers
		}(ctx, group, nil)
	}
	go func() {
		defer close(doChan)
		wg.Wait()
	}()

	AggregateErr := (error)(nil)
	for item := range doChan {
		switch assertValue := item.(type) {
		case error:
			AggregateErr = multierror.Append(AggregateErr, assertValue)
		case []*schema.GroupMember:
			result = append(result, assertValue...)
		}
	}
	if AggregateErr != nil {
		return nil, AggregateErr
	}

	return result, nil
}

func (g *gitlabIdp) GetGroupMembers(ctx context.Context, group *schema.Group, user *schema.User) ([]*schema.GroupMember, error) {
	err := g.newClient()
	if err != nil {
		return nil, fmt.Errorf("init gitlab client fail, err:【%w】", err)
	}
	page := 1
	opts := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabGroupMemberPageSize},
	}
	list, rsp, err := g.client.Groups.ListAllGroupMembers(group.Identity, opts)
	if err != nil {
		return nil, err
	}
	totalCount := rsp.TotalItems
	groupMembers := make([]*gitlab.GroupMember, 0)
	for _, groupMember := range list {
		groupMembers = append(groupMembers, groupMember)
	}
	totalPage := cast.ToInt(math.Ceil(cast.ToFloat64(totalCount) / cast.ToFloat64(gitlabGroupMemberPageSize)))
	doChan := make(chan interface{}, 1)
	wg := sync.WaitGroup{}
	for page = 2; page <= totalPage; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			opts := &gitlab.ListGroupMembersOptions{
				ListOptions: gitlab.ListOptions{Page: page, PerPage: gitlabGroupMemberPageSize},
			}
			list, _, err := g.client.Groups.ListAllGroupMembers(group.Identity, opts)
			if err != nil {
				doChan <- err
				return
			}
			doChan <- list
			return
		}(page)
	}
	go func() {
		defer close(doChan)
		wg.Wait()
	}()
	for item := range doChan {
		switch assertValue := item.(type) {
		case error:
			return nil, assertValue
		case []*gitlab.GroupMember:
			groupMembers = append(groupMembers, assertValue...)
		}
	}
	result := make([]*schema.GroupMember, 0, len(groupMembers))
	for _, groupMember := range groupMembers {
		item := g.converter.ToIdpGroupMember(group.Identity, groupMember)
		result = append(result, item)
	}
	return result, nil
}

func (g *gitlabIdp) GetProjectMembers(ctx context.Context, project *schema.Project, user *schema.User) ([]*schema.ProjectMember, error) {
	return nil, nil
}

func (g *gitlabIdp) newClient() error {
	if g.client == nil {
		// get access_token from secretProvider
		accessToken, err := g.secretProvider.GetApplicationToken(secret_provider.Identity{Type: g.Kind().Tostring(), Name: g.GetName()})
		if err != nil {
			return fmt.Errorf("get token fail, err:%w", err)
		}
		// init gitlab client
		client, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(g.apiServerUrl))
		if err != nil {
			return err
		}
		g.client = client
	}
	return nil
}
