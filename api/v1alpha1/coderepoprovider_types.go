/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	"github.com/nautes-labs/base-operator/pkg/ref_resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodeRepoProviderSpec defines the desired state of CodeRepoProvider
type CodeRepoProviderSpec struct {
	URL          string `json:"url"`
	ApiServer    string `json:"apiserver"`
	ProviderType string `json:"providertype"`
}

// CodeRepoProviderStatus defines the observed state of CodeRepoProvider
type CodeRepoProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName={coderepoprovider}
//+kubebuilder:subresource:status

// CodeRepoProvider is the Schema for the coderepoproviders API
type CodeRepoProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodeRepoProviderSpec   `json:"spec,omitempty"`
	Status CodeRepoProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CodeRepoProviderList contains a list of CodeRepoProvider
type CodeRepoProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodeRepoProvider `json:"items"`
}

func (r *CodeRepoProvider) Get(ctx context.Context, client client.Client, name, namespace string) (*ref_resource.ReferenceResourceResult, error) {
	codeRepoProvider := &CodeRepoProvider{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, codeRepoProvider)
	if err != nil {
		return nil, err
	}
	result := ref_resource.ReferenceResourceResult{
		ApiServerUrl: codeRepoProvider.Spec.ApiServer,
		ProviderType: codeRepoProvider.Spec.ProviderType,
	}
	return &result, nil
}

func init() {
	SchemeBuilder.Register(&CodeRepoProvider{}, &CodeRepoProviderList{})
}
