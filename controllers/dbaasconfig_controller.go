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
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// DBaaSConfigReconciler reconciles a DBaaSConfig object
type DBaaSConfigReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	var config v1alpha1.DBaaSConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Config for reconcile")
		return ctrl.Result{}, err
	}

	configList, err := r.configListByNS(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to list configs")
		return ctrl.Result{}, err
	}

	cond := &metav1.Condition{
		Type:    v1alpha1.DBaaSConfigReadyType,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.DBaaSConfigNotReady,
		Message: v1alpha1.MsgConfigNotReady,
	}
	if getNumActive(req.Name, configList) == 0 {
		cond = &metav1.Condition{
			Type:    v1alpha1.DBaaSConfigReadyType,
			Status:  metav1.ConditionTrue,
			Reason:  v1alpha1.Ready,
			Message: v1alpha1.MsgConfigReady,
		}
	}

	return r.updateStatusCondition(ctx, config, cond)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSConfig{}).
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

func getNumActive(name string, configList v1alpha1.DBaaSConfigList) (numActive int) {
	for i := range configList.Items {
		if name != configList.Items[i].Name &&
			apimeta.IsStatusConditionTrue(configList.Items[i].Status.Conditions, v1alpha1.DBaaSConfigReadyType) {
			numActive += 1
		}
	}
	return
}
