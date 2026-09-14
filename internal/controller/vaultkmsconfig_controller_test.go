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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kmsv1alpha1 "github.com/kevinrizza/vault-kms-plugin-openshift-provider/api/v1alpha1"
)

var _ = Describe("VaultKMSConfig Controller", func() {
	const (
		timeout          = 10 * time.Second
		interval         = 250 * time.Millisecond
		testConfigName   = "test-config"
		testVaultAddress = "https://vault.example.com:8200"
		testVaultKeyPath = "transit/keys/my-key"
		testSecretName   = "vault-approle-creds"
	)

	AfterEach(func() {
		config := &kmsv1alpha1.VaultKMSConfig{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, config)
		if err == nil {
			Expect(k8sClient.Delete(ctx, config)).To(Succeed())
		}
	})

	Context("when a VaultKMSConfig is created", func() {
		It("should populate the status with the default plugin image", func() {
			config := &kmsv1alpha1.VaultKMSConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: testConfigName,
				},
				Spec: kmsv1alpha1.VaultKMSConfigSpec{
					VaultAddress: testVaultAddress,
					VaultKeyPath: testVaultKeyPath,
					Authentication: kmsv1alpha1.VaultAuthentication{
						Type: kmsv1alpha1.VaultAuthenticationTypeAppRole,
						AppRole: kmsv1alpha1.VaultAppRoleAuthentication{
							Secret: kmsv1alpha1.VaultSecretReference{
								Name: testSecretName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, config)).To(Succeed())

			Eventually(func(g Gomega) {
				var fetched kmsv1alpha1.VaultKMSConfig
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, &fetched)).To(Succeed())
				g.Expect(fetched.Status.KMSPluginImage).To(Equal(DefaultKMSPluginImage))
			}, timeout, interval).Should(Succeed())
		})

		It("should set the default plugin image regardless of optional spec fields", func() {
			config := &kmsv1alpha1.VaultKMSConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: testConfigName,
				},
				Spec: kmsv1alpha1.VaultKMSConfigSpec{
					VaultAddress:       testVaultAddress,
					VaultKeyPath:       testVaultKeyPath,
					VaultNamespace:     "admin/team-a",
					VaultAuthNamespace: "admin/auth",
					TLS: kmsv1alpha1.VaultTLSConfig{
						CABundle: kmsv1alpha1.VaultConfigMapReference{
							Name: "vault-ca-bundle",
						},
						ServerName: "vault.internal.example.com",
					},
					Authentication: kmsv1alpha1.VaultAuthentication{
						Type: kmsv1alpha1.VaultAuthenticationTypeAppRole,
						AppRole: kmsv1alpha1.VaultAppRoleAuthentication{
							Secret: kmsv1alpha1.VaultSecretReference{
								Name: testSecretName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, config)).To(Succeed())

			Eventually(func(g Gomega) {
				var fetched kmsv1alpha1.VaultKMSConfig
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, &fetched)).To(Succeed())
				g.Expect(fetched.Status.KMSPluginImage).To(Equal(DefaultKMSPluginImage))
			}, timeout, interval).Should(Succeed())
		})

		It("should preserve the plugin image after spec changes", func() {
			config := &kmsv1alpha1.VaultKMSConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: testConfigName,
				},
				Spec: kmsv1alpha1.VaultKMSConfigSpec{
					VaultAddress: testVaultAddress,
					VaultKeyPath: testVaultKeyPath,
					Authentication: kmsv1alpha1.VaultAuthentication{
						Type: kmsv1alpha1.VaultAuthenticationTypeAppRole,
						AppRole: kmsv1alpha1.VaultAppRoleAuthentication{
							Secret: kmsv1alpha1.VaultSecretReference{
								Name: testSecretName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, config)).To(Succeed())

			Eventually(func(g Gomega) {
				var fetched kmsv1alpha1.VaultKMSConfig
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, &fetched)).To(Succeed())
				g.Expect(fetched.Status.KMSPluginImage).To(Equal(DefaultKMSPluginImage))
			}, timeout, interval).Should(Succeed())

			By("updating the spec")
			var current kmsv1alpha1.VaultKMSConfig
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, &current)).To(Succeed())
			current.Spec.VaultAddress = "https://vault-new.example.com:8200"
			current.Spec.VaultKeyPath = "transit/keys/new-key"
			Expect(k8sClient.Update(ctx, &current)).To(Succeed())

			Eventually(func(g Gomega) {
				var fetched kmsv1alpha1.VaultKMSConfig
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: testConfigName}, &fetched)).To(Succeed())
				g.Expect(fetched.Status.KMSPluginImage).To(Equal(DefaultKMSPluginImage))
			}, timeout, interval).Should(Succeed())
		})
	})
})
