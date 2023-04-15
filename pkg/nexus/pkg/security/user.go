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

package security

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nautes-labs/base-operator/pkg/nexus/pkg/client"
	"github.com/nautes-labs/base-operator/pkg/nexus/schema/security"
	"github.com/nautes-labs/base-operator/pkg/util"
)

const (
	securityUsersAPIEndpoint = securityAPIEndpoint + "/users"
	DefaultPasswd            = "123456"
	ActiveStatus             = "Active"
	DisabledStatus           = "Disabled"
)

type SecurityUserService client.Service

func NewSecurityUserService(c *client.Client) *SecurityUserService {

	s := &SecurityUserService{
		Client: c,
	}
	return s
}

func jsonUnmarshalUsers(data []byte) ([]*security.User, error) {
	var users []*security.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("could not unmarschal users: %v", err)
	}
	return users, nil
}

func (s *SecurityUserService) Create(user security.User) error {
	user.Password = DefaultPasswd
	user.Status = ActiveStatus
	ioReader, err := util.JsonMarshalInterfaceToIOReader(user)
	if err != nil {
		return err
	}

	body, resp, err := s.Client.Post(securityUsersAPIEndpoint, ioReader)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func (s *SecurityUserService) List() ([]*security.User, error) {
	body, resp, err := s.Client.Get(securityUsersAPIEndpoint, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	users, err := jsonUnmarshalUsers(body)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *SecurityUserService) Get(id string) (*security.User, error) {
	body, resp, err := s.Client.Get(fmt.Sprintf("%s?userId=%s", securityUsersAPIEndpoint, id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	users, err := jsonUnmarshalUsers(body)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.UserID == id {
			return user, nil
		}
	}

	return nil, nil
}

func (s *SecurityUserService) Update(id string, user security.User) error {
	if user.Source == "" {
		user.Source = "default"
	}
	user.Password = DefaultPasswd
	user.Status = ActiveStatus

	ioReader, err := util.JsonMarshalInterfaceToIOReader(user)
	if err != nil {
		return err
	}

	body, resp, err := s.Client.Put(fmt.Sprintf("%s/%s", securityUsersAPIEndpoint, id), ioReader)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func (s *SecurityUserService) Delete(id string) error {
	body, resp, err := s.Client.Delete(fmt.Sprintf("%s/%s", securityUsersAPIEndpoint, id))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s", string(body))
	}
	return err
}
