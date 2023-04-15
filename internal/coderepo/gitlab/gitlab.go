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

package gitlab

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	baseinterface "github.com/nautes-labs/base-operator/pkg/interface"
	nautescrd "github.com/nautes-labs/pkg/api/v1alpha1"
	nautescfg "github.com/nautes-labs/pkg/pkg/nautesconfigs"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CA_PATH = "ca/ca.crt"
)

type GitLab struct {
	gitlab.Client
	DefaultProjectName string
}

func NewProvider(token, url string, cfg nautescfg.Config) (baseinterface.ProductProvider, error) {
	return NewGitlab(token, url, cfg)
}

func NewGitlab(token, url string, cfg nautescfg.Config) (*GitLab, error) {
	apiURL := fmt.Sprintf("%s/api/v4", url)
	opts := []gitlab.ClientOptionFunc{
		gitlab.WithBaseURL(apiURL),
	}
	if strings.HasPrefix(apiURL, "https://") {
		httpClient, err := getHttpsClient()
		if err != nil {
			return nil, err
		}
		opts = append(opts, gitlab.WithHTTPClient(httpClient))
	}
	client, err := gitlab.NewClient(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get gitlab client: %w", err)
	}
	return &GitLab{
		Client:             *client,
		DefaultProjectName: cfg.Git.DefaultProductName,
	}, nil
}

func getHttpsClient() (*http.Client, error) {
	ca, err := os.ReadFile(CA_PATH)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(ca))

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}
	return client, nil
}

func (g *GitLab) GetProducts() ([]nautescrd.Product, error) {
	products := []nautescrd.Product{}
	_, resp, err := g.Projects.ListProjects(&gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
		},
		Search: &g.DefaultProjectName,
	})
	if err != nil {
		return nil, fmt.Errorf("get project list failed: %w", err)
	}

	loopTimes := (float32(resp.TotalItems) / float32(resp.ItemsPerPage)) + 1
	orderKey := "id"
	for i := float32(1); i < loopTimes; i++ {
		gitlabProjects, _, err := g.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    int(i),
				PerPage: 50,
			},
			Search:  &g.DefaultProjectName,
			OrderBy: &orderKey,
		})
		if err != nil {
			return nil, fmt.Errorf("get project list failed: %w", err)
		}
		for _, project := range gitlabProjects {
			products = append(products, nautescrd.Product{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("product-%d", project.Namespace.ID),
				},
				Spec: nautescrd.ProductSpec{
					Name:         project.Namespace.Path,
					MetaDataPath: project.SSHURLToRepo,
				},
			})
		}
	}

	newProducts := []nautescrd.Product{}
	restProducts := []nautescrd.Product{}
	for i, product := range products {
		restProducts = products[i+1:]
		isDuplicate := false
		for _, nextProduct := range restProducts {
			if product.Name == nextProduct.Name {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			newProducts = append(newProducts, product)
		}

	}

	return newProducts, nil
}

func (g *GitLab) GetProductMeta(cotx context.Context, ID string) (*baseinterface.ProductMeta, error) {
	group, resp, err := g.Groups.GetGroup(ID, nil)
	if err != nil {
		return nil, fmt.Errorf("get group info failed. code %d: %w", resp.Response.StatusCode, err)
	}

	nameWithNamespace := fmt.Sprintf("%s/%s", group.Path, g.DefaultProjectName)
	searchWithNamespace := true
	projects, resp, err := g.Projects.ListProjects(&gitlab.ListProjectsOptions{
		SearchNamespaces: &searchWithNamespace,
		Search:           &nameWithNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("get meta data failed, code %d: %w", resp.Response.StatusCode, err)

	}
	if len(projects) != 1 {
		return nil, fmt.Errorf("meta data is nil or more than one.")
	}

	return &baseinterface.ProductMeta{
		ID:     ID,
		MetaID: fmt.Sprintf("%d", projects[0].ID),
	}, nil
}
