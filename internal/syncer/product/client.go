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

package product

import (
	"context"
	"fmt"
	"strings"
	"time"

	argocrd "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/nautes-labs/base-operator/internal/syncer/productprovider"
	nautescrd "github.com/nautes-labs/pkg/api/v1alpha1"
	nautescfg "github.com/nautes-labs/pkg/pkg/nautesconfigs"

	nautesctx "github.com/nautes-labs/pkg/pkg/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CONTEXT_KEY_NAUTES_CONFIG nautesctx.ContextKey = "product.nautes.config"
	RESOURCE_NAME_CODE_REPO   string               = "repo-%s"
)

var (
	nautesProject            = "nautes"
	kubernetesDefaultService = "https://kubernetes.default.svc"
)

type appInfo struct {
	Name   string
	Health string
}

type ProductSyncer struct {
	client       client.Client
	NautesConfig nautescfg.NautesConfigs
	Rest         *rest.Config
}

func (s *ProductSyncer) Setup() error {

	err := nautescrd.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}
	err = argocrd.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}
	err = corev1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	k8sClient, err := client.New(s.Rest, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}
	s.client = k8sClient

	return nil
}

func (s *ProductSyncer) Sync(ctx context.Context, product nautescrd.Product) error {
	label := map[string]string{nautescrd.LABEL_FROM_PRODUCT: product.Name}

	cfg, err := s.NautesConfig.GetConfigByRest(s.Rest)
	if err != nil {
		return err
	}
	ctx = NewConfigContext(ctx, *cfg)

	productID, err := getProductID(product.Name)
	if err != nil {
		return err
	}

	if err := s.syncArgoProject(ctx, nautesProject); err != nil {
		return fmt.Errorf("sync argocd project failed: %w", err)
	}

	namespaceName := product.Name
	err = s.syncNamespace(ctx, namespaceName, label)
	if err != nil {
		return fmt.Errorf("sync namespace failed: %w", err)
	}

	productProvider, err := productprovider.GetProvider(ctx, CONTEXT_KEY_NAUTES_CONFIG, s.client)
	if err != nil {
		return fmt.Errorf("get product provider failed: %w", err)
	}
	productMeta, err := productProvider.GetProductMeta(ctx, productID)
	if err != nil {
		return fmt.Errorf("get product metadata failed: %w", err)
	}
	coderepoName := fmt.Sprintf(RESOURCE_NAME_CODE_REPO, productMeta.MetaID)
	err = s.syncCoderepo(ctx, coderepoName, product, label)
	if err != nil {
		return fmt.Errorf("sync coderepo failed: %w", err)
	}

	appName := product.Name
	url := product.Spec.MetaDataPath
	err = s.syncArgoApp(ctx, appName, namespaceName, url, label)
	if err != nil {
		return fmt.Errorf("sync argocd app failed: %w", err)
	}

	return nil
}

func (s *ProductSyncer) Delete(ctx context.Context, product nautescrd.Product) error {
	label := map[string]string{nautescrd.LABEL_FROM_PRODUCT: product.Name}

	cfg, err := s.NautesConfig.GetConfigByRest(s.Rest)
	if err != nil {
		return err
	}
	ctx = NewConfigContext(ctx, *cfg)

	err = s.deleteArgoApp(ctx, label)
	if err != nil {
		return fmt.Errorf("delete argocd app failed: %w", err)
	}

	err = s.deleteNamespace(ctx, label)
	if err != nil {
		return fmt.Errorf("delete namespace failed: %w", err)
	}

	return nil
}

func (s *ProductSyncer) syncArgoProject(ctx context.Context, name string) error {
	cfg, err := FromConfigContext(ctx)
	if err != nil {
		return err
	}

	namespace := cfg.Deploy.ArgoCD.Namespace
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	project := &argocrd.AppProject{}
	err = s.client.Get(ctx, key, project)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}

		project := &argocrd.AppProject{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec: argocrd.AppProjectSpec{
				SourceRepos: []string{"*"},
				Destinations: []argocrd.ApplicationDestination{
					{
						Server:    "*",
						Namespace: "*",
					},
				},
				ClusterResourceWhitelist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "*",
					},
				},
			},
		}
		log.Log.Info("nautes project not found, create new argocd project", "projectName", project.Name)

		err := s.client.Create(ctx, project)
		if err != nil {
			return err
		}
	}

	if !project.DeletionTimestamp.IsZero() {
		return fmt.Errorf("argocd project %s is terminating", project.Name)
	}

	return nil
}

func (s *ProductSyncer) syncArgoApp(ctx context.Context, name, destNamespace, url string, label map[string]string) error {
	cfg, err := FromConfigContext(ctx)
	if err != nil {
		return err
	}

	namespace := cfg.Deploy.ArgoCD.Namespace
	kustomizePath := cfg.Deploy.ArgoCD.Kustomize.DefaultPath.DefaultProject

	appList := &argocrd.ApplicationList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(label),
		client.InNamespace(namespace),
	}
	err = s.client.List(ctx, appList, listOpts...)
	if err != nil {
		return err
	}

	switch num := len(appList.Items); num {
	case 0:
		app := &argocrd.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    label,
			},
			Spec: argocrd.ApplicationSpec{
				Source: argocrd.ApplicationSource{
					RepoURL:        url,
					Path:           kustomizePath,
					TargetRevision: "HEAD",
				},
				Destination: argocrd.ApplicationDestination{
					Server:    kubernetesDefaultService,
					Namespace: destNamespace,
				},
				Project: nautesProject,
				SyncPolicy: &argocrd.SyncPolicy{
					Automated: &argocrd.SyncPolicyAutomated{
						Prune:    true,
						SelfHeal: true,
					},
				},
			},
		}

		log.FromContext(ctx).V(1).Info("create argocd app", "appName", app.Name)
		return s.client.Create(ctx, app)
	case 1:
		app := appList.Items[0]

		if !app.DeletionTimestamp.IsZero() {
			return fmt.Errorf("argocd app %s is terminating", app.Name)
		}
		if app.Spec.Source.RepoURL != url {
			app.Spec.Source.RepoURL = url
			log.FromContext(ctx).V(1).Info("update argocd app", "appName", app.Name)
			if err := s.client.Update(ctx, &app); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("too many argocd apps")
	}

	return nil
}

func (s *ProductSyncer) deleteArgoApp(ctx context.Context, label map[string]string) error {
	cfg, err := FromConfigContext(ctx)
	if err != nil {
		return err
	}

	namespace := cfg.Deploy.ArgoCD.Namespace

	appList := &argocrd.ApplicationList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(label),
		client.InNamespace(namespace),
	}
	err = s.client.List(ctx, appList, listOpts...)
	if err != nil {
		return err
	}

	errList := []error{}
	for _, app := range appList.Items {
		log.FromContext(ctx).V(1).Info("delete argocd app", "AppName", app.Name)
		err := s.client.Delete(ctx, &app)
		if err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) != 0 {
		return fmt.Errorf("%v", errList)
	}

	for i := 0; i < 2; i++ {
		appList := &argocrd.ApplicationList{}
		err := s.client.List(ctx, appList, listOpts...)
		if err != nil {
			return err
		}
		if len(appList.Items) == 0 {
			return nil
		}
		time.Sleep(time.Second * 10)
	}

	return fmt.Errorf("wait timeout exceeded")
}

func (s *ProductSyncer) syncCoderepo(ctx context.Context, name string, product nautescrd.Product, label map[string]string) error {
	cfg, err := FromConfigContext(ctx)
	if err != nil {
		return err
	}

	coderepos := &nautescrd.CodeRepoList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(label),
		client.InNamespace(product.Namespace),
	}
	err = s.client.List(ctx, coderepos, listOpts...)
	if err != nil {
		return fmt.Errorf("get code repo failed: %w", err)
	}

	switch num := len(coderepos.Items); num {
	case 0:
		coderepo := &nautescrd.CodeRepo{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: cfg.Nautes.Namespace,
				Labels:    label,
			},
			Spec: nautescrd.CodeRepoSpec{
				Product:  product.Name,
				RepoName: cfg.Git.DefaultProductName,
				URL:      product.Spec.MetaDataPath,
			},
		}
		controllerutil.SetControllerReference(&product, coderepo, scheme.Scheme)
		return s.client.Create(ctx, coderepo)
	case 1:
		coderepo := coderepos.Items[0]
		if coderepo.Spec.URL == product.Spec.MetaDataPath &&
			coderepo.Spec.Product == product.Name &&
			coderepo.Spec.RepoName == cfg.Git.DefaultProductName {
			return nil
		}
		coderepo.Spec.Product = product.Name
		coderepo.Spec.RepoName = cfg.Git.DefaultProductName
		coderepo.Spec.URL = product.Spec.MetaDataPath
		return s.client.Update(ctx, &coderepo)

	default:
		return fmt.Errorf("too many code repo")
	}
}

func (s *ProductSyncer) syncNamespace(ctx context.Context, name string, label map[string]string) error {
	labelSelector := client.MatchingLabels(label)

	nsList := &corev1.NamespaceList{}
	err := s.client.List(ctx, nsList, labelSelector)
	if err != nil {
		return err
	}

	switch num := len(nsList.Items); num {
	case 0:
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: label,
			},
		}

		log.FromContext(ctx).V(1).Info("create namespace", "NamespaceName", ns.Name)
		return s.client.Create(ctx, ns)
	case 1:
		ns := nsList.Items[0]
		if !ns.DeletionTimestamp.IsZero() {
			return fmt.Errorf("namespace %s is terminating", ns.Name)
		}
	default:
		return fmt.Errorf("too many namespaces")
	}

	return nil
}

func (s *ProductSyncer) deleteNamespace(ctx context.Context, label map[string]string) error {
	labelSelector := client.MatchingLabels(label)
	nsList := &corev1.NamespaceList{}
	err := s.client.List(ctx, nsList, labelSelector)
	if err != nil {
		return err
	}

	errList := []error{}
	for _, ns := range nsList.Items {
		log.FromContext(ctx).V(1).Info("delete namespace", "NamespaceName", ns.Name)
		err := s.client.Delete(ctx, &ns)
		if err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) != 0 {
		return fmt.Errorf("%v", errList)
	}

	for i := 0; i < 2; i++ {
		nsList := &corev1.NamespaceList{}
		err := s.client.List(ctx, nsList, labelSelector)
		if err != nil {
			return err
		}
		if len(nsList.Items) == 0 {
			return nil
		}
		time.Sleep(time.Second * 5)
	}

	return fmt.Errorf("wait timeout exceeded")
}

func getProductID(name string) (string, error) {
	parts := strings.SplitN(name, "-", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("get product id failed")
	}
	return parts[1], nil
}

func NewConfigContext(ctx context.Context, cfg nautescfg.Config) context.Context {
	return context.WithValue(ctx, CONTEXT_KEY_NAUTES_CONFIG, cfg)
}

func FromConfigContext(ctx context.Context) (*nautescfg.Config, error) {
	cfgInterface := ctx.Value(CONTEXT_KEY_NAUTES_CONFIG)
	cfg, ok := cfgInterface.(nautescfg.Config)
	if !ok {
		return nil, fmt.Errorf("can not find nautes config from context")
	}
	return &cfg, nil
}
