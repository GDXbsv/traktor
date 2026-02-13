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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/GDXbsv/traktor/api/v1alpha1"
)

var _ = Describe("SecretsRefresh Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When reconciling a resource", func() {
		var (
			secretsRefreshName string
			testNamespace      string
			deploymentName     string
			secretName         string
		)

		ctx := context.Background()

		namespace := &corev1.Namespace{}
		secretsRefresh := &appsv1alpha1.SecretsRefresh{}
		deployment := &appsv1.Deployment{}
		secret := &corev1.Secret{}

		BeforeEach(func() {
			// Generate unique names to avoid conflicts
			uniqueID := fmt.Sprintf("%d", time.Now().UnixNano())
			testNamespace = fmt.Sprintf("test-ns-%s", uniqueID)
			secretsRefreshName = fmt.Sprintf("test-sr-%s", uniqueID)
			deploymentName = fmt.Sprintf("test-deploy-%s", uniqueID)
			secretName = fmt.Sprintf("test-secret-%s", uniqueID)

			By("Creating the test namespace")
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testNamespace,
					Labels: map[string]string{
						"environment": "test",
						"team":        "platform",
					},
				},
			}
			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

			By("Creating a test deployment that uses the secret")
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:alpine",
									Env: []corev1.EnvVar{
										{
											Name: "PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: secretName,
													},
													Key: "password",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			By("Creating a test secret with labels")
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNamespace,
					Labels: map[string]string{
						"auto-refresh": "enabled",
						"type":         "app-config",
					},
				},
				StringData: map[string]string{
					"password": "initial-password",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			By("Creating the SecretsRefresh resource")
			secretsRefresh = &appsv1alpha1.SecretsRefresh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretsRefreshName,
					Namespace: "default",
				},
				Spec: appsv1alpha1.SecretsRefreshSpec{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"environment": "test",
						},
					},
					SecretSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"auto-refresh": "enabled",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, secretsRefresh)).To(Succeed())
		})

		AfterEach(func() {
			By("Cleaning up test resources")

			// Delete namespace (cascades to all resources in it)
			ns := &corev1.Namespace{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: testNamespace}, ns)
			if err == nil {
				Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
			}

			// Delete SecretsRefresh
			resource := &appsv1alpha1.SecretsRefresh{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: secretsRefreshName, Namespace: "default"}, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile and restart deployment when secret changes", func() {
			By("Reconciling with the secret name and namespace (simulating secret change)")
			controllerReconciler := &SecretsRefreshReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      secretName,
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the deployment has restart annotation")
			Eventually(func() bool {
				updatedDeployment := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName,
					Namespace: testNamespace,
				}, updatedDeployment)
				if err != nil {
					return false
				}

				annotations := updatedDeployment.Spec.Template.Annotations
				_, hasAnnotation := annotations["traktor.gdxcloud.net/restartedAt"]
				return hasAnnotation
			}, timeout, interval).Should(BeTrue())
		})

		It("should not restart deployment that doesn't use the secret", func() {
			By("Creating a second deployment without the secret")
			deploymentName2 := fmt.Sprintf("test-deploy-no-secret-%d", time.Now().UnixNano())
			deployment2 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName2,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-no-secret",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test-no-secret",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:alpine",
									// No secret reference
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment2)).To(Succeed())

			By("Reconciling with the secret name and namespace")
			controllerReconciler := &SecretsRefreshReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      secretName,
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the first deployment (with secret) has restart annotation")
			Eventually(func() bool {
				updatedDeployment := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName,
					Namespace: testNamespace,
				}, updatedDeployment)
				if err != nil {
					return false
				}

				annotations := updatedDeployment.Spec.Template.Annotations
				_, hasAnnotation := annotations["traktor.gdxcloud.net/restartedAt"]
				return hasAnnotation
			}, timeout, interval).Should(BeTrue())

			By("Verifying the second deployment (without secret) does NOT have restart annotation")
			Consistently(func() bool {
				updatedDeployment2 := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName2,
					Namespace: testNamespace,
				}, updatedDeployment2)
				if err != nil {
					return false
				}

				annotations := updatedDeployment2.Spec.Template.Annotations
				_, hasAnnotation := annotations["traktor.gdxcloud.net/restartedAt"]
				return !hasAnnotation // Should NOT have the annotation
			}, time.Second*2, interval).Should(BeTrue())
		})

		It("should restart deployment that uses secret via envFrom with multiple secrets", func() {
			By("Creating a deployment with multiple envFrom secretRefs")
			deploymentName3 := fmt.Sprintf("test-deploy-envfrom-%d", time.Now().UnixNano())
			secretName2 := fmt.Sprintf("secrets-global-%d", time.Now().UnixNano())
			secretName3 := fmt.Sprintf("secrets-terraform-%d", time.Now().UnixNano())

			deployment3 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName3,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-envfrom",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test-envfrom",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "app",
									Image: "nginx:alpine",
									EnvFrom: []corev1.EnvFromSource{
										{
											SecretRef: &corev1.SecretEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: secretName2,
												},
											},
										},
										{
											SecretRef: &corev1.SecretEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: secretName3,
												},
											},
										},
										{
											SecretRef: &corev1.SecretEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: secretName, // The original test secret
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment3)).To(Succeed())

			By("Creating additional secrets")
			secret2 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName2,
					Namespace: testNamespace,
				},
				StringData: map[string]string{
					"global_key": "global_value",
				},
			}
			Expect(k8sClient.Create(ctx, secret2)).To(Succeed())

			secret3 := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName3,
					Namespace: testNamespace,
				},
				StringData: map[string]string{
					"terraform_key": "terraform_value",
				},
			}
			Expect(k8sClient.Create(ctx, secret3)).To(Succeed())

			By("Reconciling with the middle secret (secrets-terraform)")
			controllerReconciler := &SecretsRefreshReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      secretName3, // Change secrets-terraform
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the deployment with envFrom has restart annotation")
			Eventually(func() bool {
				updatedDeployment3 := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName3,
					Namespace: testNamespace,
				}, updatedDeployment3)
				if err != nil {
					return false
				}

				annotations := updatedDeployment3.Spec.Template.Annotations
				_, hasAnnotation := annotations["traktor.gdxcloud.net/restartedAt"]
				return hasAnnotation
			}, timeout, interval).Should(BeTrue())
		})

		It("should filter namespaces correctly based on selector", func() {
			By("Getting filtered namespaces")
			controllerReconciler := &SecretsRefreshReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			namespaces, err := controllerReconciler.getFilteredNamespaces(ctx, secretsRefresh)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying test namespace is included")
			found := false
			for _, ns := range namespaces {
				if ns.Name == testNamespace {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})
	})
})

func int32Ptr(i int32) *int32 {
	return &i
}
