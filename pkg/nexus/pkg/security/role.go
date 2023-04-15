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
	"net/url"

	"github.com/nautes-labs/base-operator/pkg/nexus/pkg/client"
	"github.com/nautes-labs/base-operator/pkg/nexus/schema/security"
	"github.com/nautes-labs/base-operator/pkg/util"
)

const (
	securityrolesAPIEndpoint = securityAPIEndpoint + "/roles"
)

type SecurityRoleService client.Service

func NewSecurityRoleService(c *client.Client) *SecurityRoleService {

	s := &SecurityRoleService{
		Client: c,
	}
	return s
}

func (s *SecurityRoleService) Create(role security.Role) error {
	ioReader, err := util.JsonMarshalInterfaceToIOReader(role)
	if err != nil {
		return err
	}

	body, resp, err := s.Client.Post(securityrolesAPIEndpoint, ioReader)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func (s *SecurityRoleService) List() ([]*security.Role, error) {
	body, resp, err := s.Client.Get(securityrolesAPIEndpoint, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}
	var roles []*security.Role
	if err := json.Unmarshal(body, &roles); err != nil {
		return nil, fmt.Errorf("could not unmarshal roles: %v", err)
	}
	return roles, nil
}

func (s *SecurityRoleService) GetById(id string) (*security.Role, error) {
	encodedID := url.PathEscape(id)

	body, resp, err := s.Client.Get(fmt.Sprintf("%s/%s", securityrolesAPIEndpoint, encodedID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	var role security.Role
	if err := json.Unmarshal(body, &role); err != nil {
		return nil, fmt.Errorf("could not unmarshal roles: %v", err)
	}
	return &role, nil

}

func (s *SecurityRoleService) Update(id string, role security.Role) error {
	encodedID := url.PathEscape(id)

	ioReader, err := util.JsonMarshalInterfaceToIOReader(role)
	if err != nil {
		return err
	}

	body, resp, err := s.Client.Put(fmt.Sprintf("%s/%s", securityrolesAPIEndpoint, encodedID), ioReader)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func (s *SecurityRoleService) Delete(id string) error {
	encodedID := url.PathEscape(id)

	body, resp, err := s.Client.Delete(fmt.Sprintf("%s/%s", securityrolesAPIEndpoint, encodedID))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}
