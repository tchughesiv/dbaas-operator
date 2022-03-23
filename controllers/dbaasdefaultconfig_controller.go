/*
Copyright 2021.
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

package controllers

import (
	"context"
	"time"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DBaaSDefaultConfigReconciler reconciles a DBaaSInventory object
type DBaaSDefaultConfigReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSDefaultConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// on operator startup, create default config if none exists
	return r.createDefaultConfig(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSDefaultConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// watch deployments if installed to the operator's namespace
	return ctrl.NewControllerManagedBy(mgr).
		Named("defaultconfig").
		For(
			&appsv1.Deployment{},
			builder.WithPredicates(r.ignoreOtherDeployments()),
			builder.OnlyMetadata,
		).
		Complete(r)
}

// only reconcile deployments which reside in the operator's install namespace, and only create events
func (r *DBaaSDefaultConfigReconciler) ignoreOtherDeployments() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetNamespace() == r.InstallNamespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// create a default Config if one doesn't exist
func (r *DBaaSDefaultConfigReconciler) createDefaultConfig(ctx context.Context) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	defaultConfig := getDefaultConfig(r.InstallNamespace)

	// get list of DBaaSConfigs for install/default namespace
	configList, err := r.configListByNS(ctx, defaultConfig.Namespace)
	if err != nil {
		logger.Error(err, "unable to list configs")
		return ctrl.Result{}, err
	}

	// if no default config exists, create one
	if len(configList.Items) == 0 && !contains(getConfigNames(configList), defaultConfig.Name) {
		if err := r.Get(ctx, client.ObjectKeyFromObject(&defaultConfig), &v1alpha1.DBaaSConfig{}); err != nil {
			// proceed only if default config not found
			if errors.IsNotFound(err) {
				logger.Info("resource not found", "Name", defaultConfig.Name)
				if err := r.Create(ctx, &defaultConfig); err != nil {
					// trigger retry if creation of default config fails
					logger.Error(err, "Error creating DBaaS Config resource", "Name", defaultConfig.Name)
					return ctrl.Result{RequeueAfter: time.Duration(30) * time.Second}, err
				}
				logger.Info("creating default DBaaS Config resource", "Name", defaultConfig.Name)
			} else {
				logger.Error(err, "Error getting the DBaaS Config resource", "Name", defaultConfig.Name)
			}
		}
	}

	return ctrl.Result{}, nil
}

func getDefaultConfig(inventoryNamespace string) v1alpha1.DBaaSConfig {
	isTrue := true
	config := v1alpha1.DBaaSConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: inventoryNamespace,
		},
		Spec: v1alpha1.DBaaSConfigSpec{
			DBaaSInventoryConfigs: v1alpha1.DBaaSInventoryConfigs{
				AllowProvisions:      &isTrue,
				ConnectionNamespaces: []string{"*"},
			},
		},
	}
	config.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("DBaaSConfig"))
	return config
}

// get config names from list
func getConfigNames(configList v1alpha1.DBaaSConfigList) (configNames []string) {
	for _, config := range configList.Items {
		configNames = append(configNames, config.Name)
	}
	return
}
