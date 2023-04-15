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

package productprovider

import (
	"context"

	baseinterface "github.com/nautes-labs/base-operator/pkg/interface"
	nautescrd "github.com/nautes-labs/pkg/api/v1alpha1"
	nautescfg "github.com/nautes-labs/pkg/pkg/nautesconfigs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

var MockProvider *MockProductProvider

type MockProductProviderFactory struct{}

func (p *MockProductProviderFactory) GetProvider(ctx context.Context,
	name string,
	k8s client.Client,
	cfg nautescfg.Config) (baseinterface.ProductProvider, error) {
	return MockProvider, nil
}

type MockProductProvider struct {
	ProductList []nautescrd.Product
	ProductMeta *baseinterface.ProductMeta
}

func (p *MockProductProvider) GetProducts() ([]nautescrd.Product, error) {
	return p.ProductList, nil
}

func (p *MockProductProvider) GetProductMeta(ctx context.Context, ID string) (*baseinterface.ProductMeta, error) {
	return p.ProductMeta, nil
}
