//go:build e2e
// +build e2e

package e2e

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kmsv1alpha1 "github.com/kevinrizza/vault-kms-plugin-openshift-provider/api/v1alpha1"
)

var _ = Describe("Vault KMS Plugin OpenShift Provider", Ordered, func() {
	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("OLM installation", func() {
		It("should have a CSV in Succeeded phase", func() {
			csvList, err := kubeClient.Discovery().RESTClient().
				Get().
				AbsPath("/apis/operators.coreos.com/v1alpha1").
				Namespace(namespace).
				Resource("clusterserviceversions").
				DoRaw(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(csvList)).To(ContainSubstring(`"phase":"Succeeded"`))
		})

		It("should have the controller manager pod running", func() {
			verifyPodRunning := func(g Gomega) {
				pods, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: "control-plane=controller-manager",
				})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pods.Items).NotTo(BeEmpty())
				g.Expect(string(pods.Items[0].Status.Phase)).To(Equal("Running"))
			}
			Eventually(verifyPodRunning).Should(Succeed())
		})
	})

	Context("VaultKMSConfig CRD", func() {
		const configName = "e2e-vault-config"

		AfterAll(func() {
			config := &kmsv1alpha1.VaultKMSConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: configName,
				},
			}
			_ = k8sClient.Delete(ctx, config)
		})

		It("should accept the CRD on the cluster", func() {
			_, err := kubeClient.Discovery().RESTClient().
				Get().
				AbsPath("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/vaultkmsconfigs.kms.openshift.io").
				DoRaw(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a VaultKMSConfig resource", func() {
			config := &kmsv1alpha1.VaultKMSConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: configName,
				},
				Spec: kmsv1alpha1.VaultKMSConfigSpec{
					VaultAddress: "https://vault.example.com:8200",
					VaultKeyPath: "transit/keys/my-key",
					Authentication: kmsv1alpha1.VaultAuthentication{
						Type: kmsv1alpha1.VaultAuthenticationTypeAppRole,
						AppRole: kmsv1alpha1.VaultAppRoleAuthentication{
							Secret: kmsv1alpha1.VaultSecretReference{
								Name: "vault-approle-creds",
							},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reconcile the status with spec fields and plugin image", func() {
			verifyStatus := func(g Gomega) {
				config := &kmsv1alpha1.VaultKMSConfig{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: configName}, config)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(config.Status.VaultAddress).To(Equal("https://vault.example.com:8200"))
				g.Expect(config.Status.VaultKeyPath).To(Equal("transit/keys/my-key"))
				g.Expect(config.Status.KMSPluginImage).To(Equal("quay.io/kevinrizza/test-vault-plugin-image:latest"))
				g.Expect(string(config.Status.Authentication.Type)).To(Equal("AppRole"))
			}
			Eventually(verifyStatus).Should(Succeed())
		})

		It("should update the status when the spec is modified", func() {
			By("patching the spec with a new vault address")
			config := &kmsv1alpha1.VaultKMSConfig{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: configName}, config)
			Expect(err).NotTo(HaveOccurred())

			patch := client.MergeFrom(config.DeepCopy())
			config.Spec.VaultAddress = "https://vault-new.example.com:8200"
			err = k8sClient.Patch(ctx, config, patch)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the status reflects the new address")
			verifyUpdated := func(g Gomega) {
				updated := &kmsv1alpha1.VaultKMSConfig{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: configName}, updated)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(updated.Status.VaultAddress).To(Equal("https://vault-new.example.com:8200"))
				g.Expect(updated.Status.KMSPluginImage).To(Equal("quay.io/kevinrizza/test-vault-plugin-image:latest"))
			}
			Eventually(verifyUpdated).Should(Succeed())
		})
	})
})
