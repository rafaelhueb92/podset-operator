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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github/rafaelhueb92/podset-operator/api/v1"
)

// PodSetReconciler reconciles a PodSet object
type PodSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.podset.com,resources=podsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.podset.com,resources=podsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.podset.com,resources=podsets/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps.podset.com,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.25.0/pkg/reconcile
func (r *PodSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the Podset resource
	var podSet appsv1.PodSet
	if err := r.Get(ctx, req.NamespacedName, &podSet); err != nil {
		log.Error(err, "unable to fetch PodSet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log the reconciliation process
	log.Info("Reconciling PodSet", "name", podSet.Name, "namespace", podSet.Namespace)

	// Create or Update a configmap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podSet.Name + "-config",
			Namespace: podSet.Namespace,
		},
		Data: map[string]string{
			"replicas": fmt.Sprintf("%d", podSet.Spec.Replicas),
			"image":    podSet.Spec.Image,
			"status":   "managed-by-operator",
		},
	}

	// set Podset as owner of configmap (for garbage collection)
	if err := ctrl.SetControllerReference(&podSet, configMap, r.Scheme); err != nil {
		log.Error(err, "unable to set controller reference for ConfigMap")
		return ctrl.Result{}, err
	}

	// Create or Update the configmap
	err := r.Create(ctx, configMap)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Update the existing configmap
			err = r.Update(ctx, configMap)
			if err != nil {
				log.Error(err, "unable to update ConfigMap")
				return ctrl.Result{}, err
			}
		} else {
			log.Error(err, "unable to create ConfigMap")
			return ctrl.Result{}, err

		}
	}

	// Update the PodSet Status
	podSet.Status.Ready = true
	podSet.Status.Message = "ConfigMap Created Successfully"
	if err := r.Status().Update(ctx, &podSet); err != nil {
		log.Error(err, "Failed to update PodSet status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconcilied PodSet")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.PodSet{}).
		Named("podset").
		Complete(r)
}
