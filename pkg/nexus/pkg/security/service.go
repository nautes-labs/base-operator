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

import "github.com/nautes-labs/base-operator/pkg/nexus/pkg/client"

const (
	securityAPIEndpoint = client.BasePath + "v1/security"
)

type SecurityService struct {
	client *client.Client
	Role   *SecurityRoleService
	User   *SecurityUserService
}

func NewSecurityService(c *client.Client) *SecurityService {
	return &SecurityService{
		client: c,
		Role:   NewSecurityRoleService(c),
		User:   NewSecurityUserService(c),
	}
}
