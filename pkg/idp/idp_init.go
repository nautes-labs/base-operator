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
	"fmt"
	reflect "reflect"
)

var (
	IdpKindMapping = make(map[string]Idp, 0)
)

func init() {
	IdpKindMapping[GitlabIdpKind.Tostring()] = (*gitlabIdp)(nil)
}

func NewIdp(idpKind string) (Idp, error) {
	idpEntity, ok := IdpKindMapping[idpKind]
	if !ok {
		return nil, fmt.Errorf("platform unsupport idp type:%s", idpKind)
	}
	t := reflect.TypeOf(idpEntity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(Idp), nil
}
