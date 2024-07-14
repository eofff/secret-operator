/*
Copyright 2024.

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
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretv1 "github.com/eofff/secret-operator/api/v1"
)

// SecretWatcherReconciler reconciles a SecretWatcher object
type SecretWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=secret.fearning,resources=secretwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secret.fearning,resources=secretwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=secret.fearning,resources=secretwatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SecretWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	secretWatcher := &secretv1.SecretWatcher{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		secretWatcher,
	)

	if err != nil {
		return ctrl.Result{}, err
	}

	for _, ns := range secretWatcher.Spec.Namespaces {
		namespace := &v1.Namespace{}
		err = r.Get(
			ctx,
			types.NamespacedName{
				Name:      ns,
				Namespace: ns,
			},
			namespace,
		)

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	copiedSecrets := &v1.SecretList{}
	copiedSecretOpts := []client.ListOption{
		client.MatchingLabels{"copied": "true"},
	}
	err = r.List(ctx, copiedSecrets, copiedSecretOpts...)

	if err != nil {
		return ctrl.Result{}, err
	}

	unusedCopied := make(map[types.NamespacedName]bool)
	for _, secret := range copiedSecrets.Items {
		unusedCopied[types.NamespacedName{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		}] = true
	}

	copymeSecrets := &v1.SecretList{}
	copymeSecretOpts := []client.ListOption{
		client.MatchingLabels{"copy-me": "true"},
	}
	err = r.List(ctx, copymeSecrets, copymeSecretOpts...)

	if err != nil {
		return ctrl.Result{}, err
	}

	for _, originalSecret := range copymeSecrets.Items {
		headlessOriginal := originalSecret
		delete(headlessOriginal.Labels, "copy-me")
		headlessOriginal.Namespace = "buff"

		for _, ns := range secretWatcher.Spec.Namespaces {
			secret := &v1.Secret{}
			err = r.Get(ctx,
				types.NamespacedName{
					Name:      originalSecret.Name,
					Namespace: ns,
				},
				secret,
			)

			if err != nil {
				newSecret := originalSecret
				newSecret.Namespace = ns
				delete(newSecret.Labels, "copy-me")
				newSecret.Labels["copied"] = "true"
				newSecret.ResourceVersion = ""
				err = r.Create(ctx, &newSecret)
				if err == nil {
					log.Log.Info("copy of %s created in ns: %s", originalSecret.Name, ns)
				} else {
					return ctrl.Result{}, err
				}
			} else {
				delete(unusedCopied, types.NamespacedName{
					Name:      secret.Name,
					Namespace: secret.Namespace,
				})

				headlessSecret := secret
				delete(headlessSecret.Labels, "copied")
				headlessSecret.Namespace = "buff"

				if !compareSecrets(&headlessOriginal, headlessSecret) {
					updatedSecret := secret
					updatedSecret.Namespace = ns
					updatedSecret.Labels = make(map[string]string)
					updatedSecret.Data = make(map[string][]byte)

					for k, v := range originalSecret.Labels {
						if k == "copy-me" {
							continue
						}

						updatedSecret.Labels[k] = v
					}

					for k, v := range originalSecret.Data {
						updatedSecret.Data[k] = v
					}

					updatedSecret.Labels["copied"] = "true"
					err = r.Update(ctx, updatedSecret)
					if err == nil {
						log.Log.Info("copy of %s updated in ns: %s", originalSecret.Name, ns)
					} else {
						return ctrl.Result{}, err
					}
				}
			}
		}
	}

	for secretTag := range unusedCopied {
		secret := &v1.Secret{}
		err = r.Get(ctx, secretTag, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = r.Delete(ctx, secret)
		if err == nil {
			log.Log.Info("copy of %s deleted in ns: %s", secretTag.Name, secretTag.Namespace)
		} else {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{
		Requeue:      false,
		RequeueAfter: time.Duration(int64(secretWatcher.Spec.CheckTimeoutSeconds)) * time.Second,
	}, nil
}

func compareSecrets(secret1, secret2 *v1.Secret) bool {
	for label1, value1 := range secret1.Labels {
		if label1 == "copy-me" || label1 == "copied" {
			continue
		}

		labelFinded := false
		for label2, value2 := range secret2.Labels {
			if label1 == label2 {
				labelFinded = value1 == value2
				break
			}
		}
		if !labelFinded {
			return false
		}
	}

	for field1, value1 := range secret1.Data {
		goodCompare := false

		for field2, value2 := range secret2.Data {
			if field1 == field2 && len(value1) == len(value2) {
				for i := 0; i < len(value1); i++ {
					if value1[i] != value2[i] {
						return false
					}
				}
			}
		}

		if !goodCompare {
			return false
		}
	}

	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretv1.SecretWatcher{}).
		Complete(r)
}
