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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DBaaSConfigReconciler reconciles a DBaaSConfig object
type DBaaSConfigReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=core,resources=resourcequotas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	configList, err := r.configListByNS(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to list configs")
		return ctrl.Result{}, err
	}
	activeConfig := getActiveConfig(configList)

	var config v1alpha1.DBaaSConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued
			if len(configList.Items) > 0 && activeConfig == nil {
				// reconcile another config to ensure one is active
				config = configList.Items[0]
			} else {
				// child objects getting GC'd, no requeue
				return ctrl.Result{}, nil
			}
		} else {
			logger.Error(err, "Error fetching DBaaS Config for reconcile")
			return ctrl.Result{}, err
		}
	}
	cond := &metav1.Condition{
		Type:    v1alpha1.DBaaSConfigReadyType,
		Status:  metav1.ConditionTrue,
		Reason:  v1alpha1.Ready,
		Message: v1alpha1.MsgConfigReady,
	}

	// if an active config exists, and it's not this one... set status to false
	if activeConfig != nil &&
		activeConfig.GetName() != config.Name {
		cond = &metav1.Condition{
			Type:    v1alpha1.DBaaSConfigReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  v1alpha1.DBaaSConfigNotReady,
			Message: v1alpha1.MsgConfigNotReady + " - " + activeConfig.GetName(),
		}
	}

	// if config is active, create resourcequota
	if cond.Status == metav1.ConditionTrue {
		resQuota := v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-" + config.Name,
				Namespace: config.Namespace,
			},
		}
		if res, err := controllerutil.CreateOrUpdate(ctx, r.Client, &resQuota, func() error {
			resQuota.Spec = v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceName("count/dbaasconfigs." + v1alpha1.GroupVersion.Group): resource.MustParse("1"),
				},
			}
			resQuota.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("ResourceQuota"))
			if err := ctrl.SetControllerReference(&config, &resQuota, r.Scheme); err != nil {
				return err
			}
			return nil
		}); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("ResourceQuota resource modified, retry syncing status", "ResourceQuota", resQuota)
				return ctrl.Result{Requeue: true}, err
			}
			logger.Error(err, "Error updating the ResourceQuota resource status", "ResourceQuota", resQuota)
			return ctrl.Result{}, err
		} else if res != controllerutil.OperationResultNone {
			logger.Info("ResourceQuota resource reconciled", "ResourceQuota", resQuota, "result", res)
		}
	}

	return r.updateStatusCondition(ctx, config, cond)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSConfig{}).
		Owns(&v1.ResourceQuota{}).
		Complete(r)
}

func (r *DBaaSConfigReconciler) updateStatusCondition(ctx context.Context, config v1alpha1.DBaaSConfig, cond *metav1.Condition) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	apimeta.SetStatusCondition(&config.Status.Conditions, *cond)
	if err := r.Client.Status().Update(ctx, &config); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("DBaaS Config resource modified, retry syncing status", "DBaaS Config", config)
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Error updating the DBaaS Config resource status", "DBaaS Config", config)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// get active config, return nil if none exists
func getActiveConfig(configList v1alpha1.DBaaSConfigList) *v1alpha1.DBaaSConfig {
	for i := range configList.Items {
		if apimeta.IsStatusConditionTrue(configList.Items[i].Status.Conditions, v1alpha1.DBaaSConfigReadyType) {
			return &configList.Items[i]
		}
	}
	return nil
}
