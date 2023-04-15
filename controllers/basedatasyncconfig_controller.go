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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/nautes-labs/base-operator/pkg/idp"
	"github.com/nautes-labs/base-operator/pkg/services"
	"github.com/nautes-labs/base-operator/pkg/target"

	"github.com/nautes-labs/base-operator/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nautes-labs/base-operator/api/v1alpha1"
	nautesv1alpha1 "github.com/nautes-labs/base-operator/api/v1alpha1"
	"github.com/nautes-labs/base-operator/pkg/ref_resource"
	"github.com/nautes-labs/base-operator/pkg/secret_provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BaseDataSyncConfigReconciler reconciles a BaseDataSyncConfig object
type BaseDataSyncConfigReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	SecretProvider *secret_provider.SecretProvider
}

var (
	refResourceGvkMapping = make(map[string]ref_resource.ReferenceResource)
)

func init() {
	codeRepoProviderGvk := &metav1.GroupVersionKind{
		Group:   nautesv1alpha1.GroupVersion.Group,
		Version: nautesv1alpha1.GroupVersion.Version,
		Kind:    reflect.TypeOf(nautesv1alpha1.CodeRepoProvider{}).Name(),
	}
	refResourceGvkMapping[codeRepoProviderGvk.String()] = (*nautesv1alpha1.CodeRepoProvider)(nil)
	artifactRepoProviderGvk := &metav1.GroupVersionKind{
		Group:   nautesv1alpha1.GroupVersion.Group,
		Version: nautesv1alpha1.GroupVersion.Version,
		Kind:    reflect.TypeOf(nautesv1alpha1.ArtifactRepoProvider{}).Name(),
	}
	refResourceGvkMapping[artifactRepoProviderGvk.String()] = (*nautesv1alpha1.ArtifactRepoProvider)(nil)
}

//+kubebuilder:rbac:groups=nautes.resource.nautes.io,resources=basedatasyncconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nautes.resource.nautes.io,resources=basedatasyncconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nautes.resource.nautes.io,resources=basedatasyncconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *BaseDataSyncConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Loger.Info("trigger Reconcile")
	baseCfg := v1alpha1.BaseDataSyncConfig{}
	if err := r.Get(ctx, req.NamespacedName, &baseCfg); err != nil {
		log.Loger.Errorf("unable to fetch BaseDataSyncConfig, err:%v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	svc := services.NewSyncLogicService(ctx)
	idp, err := r.getIdpEntityByCR(ctx, baseCfg)
	if err != nil {
		log.Loger.Errorf("unable match idp, err:%v", err)
		return ctrl.Result{}, err
	}
	svc.InjectIdp(idp)
	targetApps, err := r.getTargetEntitiesByCR(ctx, idp, baseCfg)
	if err != nil {
		log.Loger.Errorf("unable match targetApp, err:%v", err)
		return ctrl.Result{}, err
	}
	svc.InjectTargetApps(targetApps...)

	// Emptying Finalizers
	// Clean up cr configured target applications
	if !baseCfg.DeletionTimestamp.IsZero() {
		err = svc.ClearTargetAppData()
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	err = svc.Run()
	if err != nil {
		return ctrl.Result{}, err
	}

	//TODO add/update Finalizers

	//TODO update k8s status

	// Add next queue consumption interval
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BaseDataSyncConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nautesv1alpha1.BaseDataSyncConfig{}).
		Complete(r)
}

// Get idp object by BaseDataSyncConfig CR
// If both of spec and ref,  spec priority is greater than ref
func (r *BaseDataSyncConfigReconciler) getIdpEntityByCR(ctx context.Context, baseCfg v1alpha1.BaseDataSyncConfig) (idp.Idp, error) {
	err := (error)(nil)
	idpApiServerUrl := ""
	idpKind := ""
	idpName := ""
	if baseCfg.Spec.Source.ApplicationSpec != nil {
		idpApiServerUrl = baseCfg.Spec.Source.ApplicationSpec.ApiServerUrl
		idpName = baseCfg.Spec.Source.ApplicationSpec.Name
		idpKind = baseCfg.Spec.Source.ApplicationSpec.ProviderType
	} else {
		refResourceResult := (*ref_resource.ReferenceResourceResult)(nil)
		refResourceResult, err = r.getRefResourceResult(ctx, baseCfg.Spec.Source.ApplicationRef)
		if err != nil {
			return nil, err
		}
		idpKind = refResourceResult.ProviderType
		idpApiServerUrl = refResourceResult.ApiServerUrl
		idpName = baseCfg.Spec.Source.ApplicationRef.Name
	}
	idpApp, err := idp.NewIdp(idpKind)
	if err != nil {
		return nil, err
	}
	idpApp.SetName(idpName)
	idpApp.SetApiServerUrl(idpApiServerUrl)
	idpApp.SetSecretProvider(r.SecretProvider)
	return idpApp, nil
}

// Get targetApp objects by BaseDataSyncConfig CR
// If both of spec and ref,  spec priority is greater than ref
func (r *BaseDataSyncConfigReconciler) getTargetEntitiesByCR(ctx context.Context, idp idp.Idp, baseCfg v1alpha1.BaseDataSyncConfig) ([]target.TargetApp, error) {
	result := make([]target.TargetApp, 0, len(baseCfg.Spec.Targets))
	err := (error)(nil)
	for _, targetCfg := range baseCfg.Spec.Targets {
		targetAppName := ""
		targetKind := ""
		apiServerUrl := ""
		if targetCfg.ApplicationSpec != nil {
			targetAppName = targetCfg.ApplicationSpec.Name
			targetKind = targetCfg.ApplicationSpec.ProviderType
			apiServerUrl = targetCfg.ApplicationSpec.ApiServerUrl
		} else {
			refResourceResult := (*ref_resource.ReferenceResourceResult)(nil)
			refResourceResult, err = r.getRefResourceResult(ctx, targetCfg.ApplicationRef)
			if err != nil {
				return nil, err
			}
			targetAppName = targetCfg.ApplicationRef.Name
			targetKind = refResourceResult.ProviderType
			apiServerUrl = refResourceResult.ApiServerUrl
		}
		targetApp, err := target.NewTargetApplication(targetKind)
		if err != nil {
			return nil, err
		}
		targetApp.SetIdp(idp)
		targetApp.SetName(targetAppName)
		targetApp.SetApiServerUrl(apiServerUrl)
		targetApp.SetSecretProvider(r.SecretProvider)
		result = append(result, targetApp)
	}
	return result, nil
}

// get reference k8s resource result
func (r *BaseDataSyncConfigReconciler) getRefResourceResult(ctx context.Context, appRef *nautesv1alpha1.ApplicationRef) (*ref_resource.ReferenceResourceResult, error) {
	refReourceGvk := schema.GroupVersionKind{
		Group:   appRef.Group,
		Version: appRef.Version,
		Kind:    appRef.Kind,
	}
	emptyRefResource, ok := refResourceGvkMapping[refReourceGvk.String()]
	if !ok {
		return nil, fmt.Errorf("unsuported k8s resources gvk:%v", refReourceGvk)
	}
	refResourceInstance := ref_resource.NewReferenceResource(emptyRefResource)
	result, err := refResourceInstance.Get(ctx, r.Client, appRef.Name, appRef.Namespace)
	if err != nil {
		return nil, err
	}
	return result, nil
}
