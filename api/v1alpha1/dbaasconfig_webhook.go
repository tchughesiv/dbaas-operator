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
	"fmt"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dbaasconfiglog = logf.Log.WithName("dbaasconfig-resource")

var configWebhookApiClient client.Client

func (r *DBaaSConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if configWebhookApiClient == nil {
		configWebhookApiClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-dbaas-redhat-com-v1alpha1-dbaasconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.redhat.com,resources=dbaasconfigs,verbs=create;update,versions=v1alpha1,name=vdbaasconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &DBaaSConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSConfig) ValidateCreate() error {
	dbaasconfiglog.Info("validate create", "name", r.Name)
	return r.validateConfig()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSConfig) ValidateUpdate(old runtime.Object) error {
	// dbaasconfiglog.Info("validate update", "name", r.Name)
	// return r.validateConfig()
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSConfig) ValidateDelete() error {
	// dbaasconfiglog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *DBaaSConfig) validateConfig() error {
	configsList := DBaaSConfigList{}
	if err := configWebhookApiClient.List(context.TODO(), &configsList, &client.ListOptions{Namespace: r.Namespace}); err != nil {
		return err
	}
	for _, config := range configsList.Items {
		if apimeta.IsStatusConditionTrue(config.Status.Conditions, DBaaSConfigReadyType) {
			errMsg := fmt.Sprintf("the namespace %s is already managed by config %s, it cannot be managed by another config", r.Namespace, config.Name)
			return field.Invalid(field.NewPath("metadata").Child("Namespace"), r.Namespace, errMsg)
		}
	}
	if len(configsList.Items) > 0 {
		errMsg := fmt.Sprintf("the namespace %s is already managed by another config", r.Namespace)
		return field.Invalid(field.NewPath("metadata").Child("Namespace"), r.Namespace, errMsg)
	}

	return nil
}
