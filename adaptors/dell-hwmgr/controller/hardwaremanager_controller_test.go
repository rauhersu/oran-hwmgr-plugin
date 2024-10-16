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

	hwmgrpluginoranopenshiftiov1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
)

var _ = Describe("HardwareManager Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		hardwaremanager := &hwmgrpluginoranopenshiftiov1alpha1.HardwareManager{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind HardwareManager")
			err := k8sClient.Get(ctx, typeNamespacedName, hardwaremanager)
			if err != nil && errors.IsNotFound(err) {
				resource := &hwmgrpluginoranopenshiftiov1alpha1.HardwareManager{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: hwmgrpluginoranopenshiftiov1alpha1.HardwareManagerSpec{
						AdaptorID: "dell-hwmgr",
						DellData: &hwmgrpluginoranopenshiftiov1alpha1.DellData{
							User: "test",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &hwmgrpluginoranopenshiftiov1alpha1.HardwareManager{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HardwareManager")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &HardwareManagerReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				Logger:    logger,
				Namespace: "default", // TODO(user):Modify as needed
				AdaptorID: "dell-hwmgr",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})