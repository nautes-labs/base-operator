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

package ref_resource

import (
	"context"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReferenceResourceResult struct {
	ApiServerUrl string
	ProviderType string
}

type ReferenceResource interface {
	Get(ctx context.Context, client client.Client, name, namespace string) (*ReferenceResourceResult, error)
}

func NewReferenceResource(referenceResource ReferenceResource) ReferenceResource {
	t := reflect.TypeOf(referenceResource)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(ReferenceResource)
}
