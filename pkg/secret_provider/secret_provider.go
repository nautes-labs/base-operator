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

package secret_provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/nautes-labs/base-operator/pkg/log"
)

type AuthenticationType string

const (
	TokenType     AuthenticationType = "token"
	BasicAuthType AuthenticationType = "basic-auth"
)

type SecretProvider struct {
	AuthenticationEntities      []AuthenticationEntity
	AuthenticationEntityMapping map[AuthenticationType]map[Identity]AuthenticationEntity
}

type AuthenticationEntity struct {
	Identity           Identity           `json:"identity"`
	AuthenticationType AuthenticationType `json:"authentication_type"`
	AuthenticationData AuthenticationData `json:"authentication_data"`
}

type Identity struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type AuthenticationData struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
}

func NewSecretProvider(filePath string) (*SecretProvider, error) {
	provider := &SecretProvider{
		make([]AuthenticationEntity, 0),
		make(map[AuthenticationType]map[Identity]AuthenticationEntity, 0),
	}
	err := provider.parseContentByPath(filePath)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (o *SecretProvider) parseContentByPath(filePath string) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Loger.Errorf("read secret file fail, err:%v", err)
		return err
	}
	err = json.Unmarshal(b, &o.AuthenticationEntities)
	if err != nil {
		log.Loger.Errorf("unserialize secret file content fail, err:%v", err)
		return err
	}
	AuthenticationTypeMapping := make(map[AuthenticationType][]AuthenticationEntity, 0)
	for _, item := range o.AuthenticationEntities {
		AuthenticationTypeMapping[item.AuthenticationType] = append(AuthenticationTypeMapping[item.AuthenticationType], item)
	}
	for authenticationType, authenticationEntities := range AuthenticationTypeMapping {
		IdentityMapping := make(map[Identity]AuthenticationEntity, 0)
		for _, authenticationEntity := range authenticationEntities {
			IdentityMapping[authenticationEntity.Identity] = authenticationEntity
		}
		o.AuthenticationEntityMapping[authenticationType] = IdentityMapping
	}
	return nil
}

func (o *SecretProvider) GetApplicationToken(Identity Identity) (token string, err error) {
	IdentityMapping, ok := o.AuthenticationEntityMapping[TokenType]
	if !ok {
		return "", fmt.Errorf("tokenType secret data is empty")
	}
	item, ok := IdentityMapping[Identity]
	if !ok {
		return "", fmt.Errorf("unkown Identity:%s authType:%s", item, TokenType)
	}
	return item.AuthenticationData.Token, nil
}

func (o *SecretProvider) GetApplicationBasicAuth(Identity Identity) (username, passwd string, err error) {
	IdentityMapping, ok := o.AuthenticationEntityMapping[BasicAuthType]
	if !ok {
		return "", "", fmt.Errorf("BasicAuthType secret data is empty")
	}
	item, ok := IdentityMapping[Identity]
	if !ok {
		return "", "", fmt.Errorf("unkown Identity:%s authType:%s", item, BasicAuthType)
	}
	return item.AuthenticationData.Username, item.AuthenticationData.Passwd, nil
}
