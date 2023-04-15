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

var _ ref_resource.ReferenceResource = (*ArtifactRepoProvider)(nil)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArtifactRepoProviderSpec defines the desired state of ArtifactRepoProvider
type ArtifactRepoProviderSpec struct {
	URL          string `json:"url"`
	ApiServer    string `json:"apiserver"`
	ProviderType string `json:"providertype"`
}

// ArtifactRepoProviderStatus defines the observed state of ArtifactRepoProvider
type ArtifactRepoProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ArtifactRepoProvider is the Schema for the artifactrepoproviders API
type ArtifactRepoProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArtifactRepoProviderSpec   `json:"spec,omitempty"`
	Status ArtifactRepoProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArtifactRepoProviderList contains a list of ArtifactRepoProvider
type ArtifactRepoProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArtifactRepoProvider `json:"items"`
}

func (r *ArtifactRepoProvider) Get(ctx context.Context, client client.Client, name, namespace string) (*ref_resource.ReferenceResourceResult, error) {
	artifactRepoProvider := &ArtifactRepoProvider{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, artifactRepoProvider)
	if err != nil {
		return nil, err
	}
	result := ref_resource.ReferenceResourceResult{
		ApiServerUrl: artifactRepoProvider.Spec.ApiServer,
		ProviderType: artifactRepoProvider.Spec.ProviderType,
	}
	return &result, nil
}

func init() {
	SchemeBuilder.Register(&ArtifactRepoProvider{}, &ArtifactRepoProviderList{})
}
