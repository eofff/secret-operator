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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretv1 "github.com/eofff/secret-operator/api/v1"

	v1core "k8s.io/api/core/v1"
)

var _ = Describe("SecretWatcher Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		secretwatcher := &secretv1.SecretWatcher{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind SecretWatcher")
			err := k8sClient.Get(ctx, typeNamespacedName, secretwatcher)
			if err != nil && errors.IsNotFound(err) {
				resource := &secretv1.SecretWatcher{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				resource.Spec.CheckTimeoutSeconds = 5
				resource.Spec.Namespaces = make([]string, 0)
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance SecretWatcher")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &SecretWatcherReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should error reconcile", func() {
			By("not existing namespaces")

			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			resource.Spec.Namespaces = append(resource.Spec.Namespaces, "testnamespace")
			err = k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &SecretWatcherReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
		})
		It("should successfully reconcile the resource when default work", func() {
			By("default work")

			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			resource.Spec.Namespaces = make([]string, 0)
			resource.Spec.Namespaces = append(resource.Spec.Namespaces, "testnamespace")
			err = k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			namespace := &v1core.Namespace{}
			namespace.Name = "testnamespace"
			namespace.Namespace = "testnamespace"
			err = k8sClient.Create(ctx, namespace)
			Expect(err).NotTo(HaveOccurred())

			secret := &v1core.Secret{}
			secret.Data = make(map[string][]byte)
			secret.Labels = make(map[string]string)
			secret.Name = "secret1"
			secret.Data["f"] = []byte("kek")
			secret.Namespace = "default"
			secret.Labels["copy-me"] = "true"
			err = k8sClient.Create(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &SecretWatcherReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			createdSecret := &v1core.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "secret1",
				Namespace: "testnamespace",
			}, createdSecret)
			Expect(err).NotTo(HaveOccurred())

			Expect(createdSecret.Labels["copied"] == "true").To(BeTrue())
			Expect(len(createdSecret.Data["f"]) == len([]byte("kek"))).To(BeTrue())
			for i := 0; i < len(createdSecret.Data["f"]); i++ {
				Expect(createdSecret.Data["f"][i] == []byte("kek")[i]).To(BeTrue())
			}

			Expect(k8sClient.Delete(ctx, namespace)).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, secret)).NotTo(HaveOccurred())
		})
		It("should successfully reconcile the resource when update", func() {
			By("default update")

			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			resource.Spec.Namespaces = make([]string, 0)
			resource.Spec.Namespaces = append(resource.Spec.Namespaces, "testnamespace2")
			err = k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			namespace := &v1core.Namespace{}
			namespace.Name = "testnamespace2"
			namespace.Namespace = "testnamespace2"
			err = k8sClient.Create(ctx, namespace)
			Expect(err).NotTo(HaveOccurred())

			secret := &v1core.Secret{}
			secret.Data = make(map[string][]byte)
			secret.Labels = make(map[string]string)
			secret.Name = "secret2"
			secret.Data["f"] = []byte("kek")
			secret.Namespace = "default"
			secret.Labels["copy-me"] = "true"
			err = k8sClient.Create(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &SecretWatcherReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			secret.Data["f"] = []byte("boo")
			err = k8sClient.Update(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			createdSecret := &v1core.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "secret2",
				Namespace: "testnamespace2",
			}, createdSecret)
			Expect(err).NotTo(HaveOccurred())

			Expect(createdSecret.Labels["copied"] == "true").To(BeTrue())
			Expect(len(createdSecret.Data["f"]) == len([]byte("boo"))).To(BeTrue())
			for i := 0; i < len(createdSecret.Data["f"]); i++ {
				Expect(createdSecret.Data["f"][i] == []byte("boo")[i]).To(BeTrue())
			}

			Expect(k8sClient.Delete(ctx, namespace)).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, secret)).NotTo(HaveOccurred())
		})
		It("should successfully reconcile the resource when delete", func() {
			By("default update")

			resource := &secretv1.SecretWatcher{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			resource.Spec.Namespaces = make([]string, 0)
			resource.Spec.Namespaces = append(resource.Spec.Namespaces, "testnamespace3")
			err = k8sClient.Update(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			namespace := &v1core.Namespace{}
			namespace.Name = "testnamespace3"
			namespace.Namespace = "testnamespace3"
			err = k8sClient.Create(ctx, namespace)
			Expect(err).NotTo(HaveOccurred())

			secret := &v1core.Secret{}
			secret.Data = make(map[string][]byte)
			secret.Labels = make(map[string]string)
			secret.Name = "secret3"
			secret.Data["f"] = []byte("kek")
			secret.Namespace = "default"
			secret.Labels["copy-me"] = "true"
			err = k8sClient.Create(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			controllerReconciler := &SecretWatcherReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Delete(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			createdSecret := &v1core.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "secret3",
				Namespace: "testnamespace3",
			}, createdSecret)
			Expect(err).To(HaveOccurred())

			Expect(k8sClient.Delete(ctx, namespace)).NotTo(HaveOccurred())
		})
	})
})
