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
	"fmt"
	"reflect"
)

var (
	AppKindMapping = make(map[string]TargetApp, 0)
)

func init() {
	AppKindMapping[string(NexusAppKind)] = (*nexusApp)(nil)
}

func NewTargetApplication(appKind string) (TargetApp, error) {
	appEntity, ok := AppKindMapping[appKind]
	if !ok {
		return nil, fmt.Errorf("platform unsupport targetApp type:%s", appKind)
	}
	t := reflect.TypeOf(appEntity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(TargetApp), nil
}
