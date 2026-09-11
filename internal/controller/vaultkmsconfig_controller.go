/*
Copyright 2026.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kmsv1alpha1 "github.com/kevinrizza/vault-kms-plugin-openshift-provider/api/v1alpha1"
)

const (
	DefaultKMSPluginImage = "quay.io/kevinrizza/test-vault-plugin-image:latest"
)

// +kubebuilder:rbac:groups=kms.openshift.io,resources=vaultkmsconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kms.openshift.io,resources=vaultkmsconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kms.openshift.io,resources=vaultkmsconfigs/finalizers,verbs=update

type VaultKMSConfigReconciler struct {
	client.Client
}

func (r *VaultKMSConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var config kmsv1alpha1.VaultKMSConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	desired := kmsv1alpha1.VaultKMSConfigStatus{
		KMSPluginImage:     DefaultKMSPluginImage,
		VaultAddress:       config.Spec.VaultAddress,
		VaultNamespace:     config.Spec.VaultNamespace,
		VaultAuthNamespace: config.Spec.VaultAuthNamespace,
		TLS:                config.Spec.TLS,
		Authentication:     config.Spec.Authentication,
		VaultKeyPath:       config.Spec.VaultKeyPath,
	}

	if !equality.Semantic.DeepEqual(config.Status, desired) {
		logger.Info("updating VaultKMSConfig status")
		config.Status = desired
		if err := r.Status().Update(ctx, &config); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *VaultKMSConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kmsv1alpha1.VaultKMSConfig{}).
		Complete(r)
}
