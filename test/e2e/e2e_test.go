//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kevinrizza/vault-kms-plugin-openshift-provider/test/utils"
)

var _ = Describe("Vault KMS Plugin OpenShift Provider", Ordered, func() {
	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("OLM installation", func() {
		It("should have a CSV in Succeeded phase", func() {
			cmd := exec.Command("kubectl", "get", "csv",
				"-n", namespace,
				"-o", "jsonpath={.items[0].status.phase}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("Succeeded"))
		})

		It("should have the controller manager pod running", func() {
			verifyPodRunning := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "pods",
					"-l", "control-plane=controller-manager",
					"-n", namespace,
					"-o", "jsonpath={.items[0].status.phase}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"))
			}
			Eventually(verifyPodRunning).Should(Succeed())
		})
	})

	Context("Namespace restriction", func() {
		const wrongNamespace = "wrong-namespace-test"

		BeforeAll(func() {
			cmd := exec.Command("kubectl", "create", "namespace", wrongNamespace)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			cmd := exec.Command("kubectl", "delete", "namespace", wrongNamespace, "--ignore-not-found")
			_, _ = utils.Run(cmd)
		})

		It("should fail to start when running in the wrong namespace", func() {
			By("getting the operator image from the existing deployment")
			cmd := exec.Command("kubectl", "get", "deployment",
				"-l", "control-plane=controller-manager",
				"-n", namespace,
				"-o", "jsonpath={.items[0].spec.template.spec.containers[0].image}")
			image, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(image).NotTo(BeEmpty())

			By("creating a pod in the wrong namespace with POD_NAMESPACE set via the downward API")
			podManifest := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: namespace-restriction-test
  namespace: %s
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: manager
    image: %s
    imagePullPolicy: IfNotPresent
    command: ["/manager"]
    env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
`, wrongNamespace, image)

			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(podManifest)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the pod fails with a namespace restriction error")
			verifyPodFailed := func(g Gomega) {
				cmd := exec.Command("kubectl", "logs",
					"namespace-restriction-test",
					"-n", wrongNamespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("Namespace restriction violated"))
			}
			Eventually(verifyPodFailed).Should(Succeed())
		})
	})

	Context("ConfigMap reconciliation", func() {
		It("should create the ConfigMap with the correct data", func() {
			verifyConfigMap := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "configmap",
					"ibm-kms-vault-plugin-provider",
					"-n", namespace,
					"-o", "jsonpath={.data.image}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("quay.io/kevinrizza/test-vault-plugin-image:latest"))
			}
			Eventually(verifyConfigMap).Should(Succeed())
		})

		It("should restore the ConfigMap after deletion", func() {
			By("deleting the ConfigMap")
			cmd := exec.Command("kubectl", "delete", "configmap",
				"ibm-kms-vault-plugin-provider",
				"-n", namespace)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the ConfigMap is recreated")
			verifyRecreated := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "configmap",
					"ibm-kms-vault-plugin-provider",
					"-n", namespace,
					"-o", "jsonpath={.data.image}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("quay.io/kevinrizza/test-vault-plugin-image:latest"))
			}
			Eventually(verifyRecreated).Should(Succeed())
		})

		It("should restore the ConfigMap data after modification", func() {
			By("modifying the ConfigMap data")
			cmd := exec.Command("kubectl", "patch", "configmap",
				"ibm-kms-vault-plugin-provider",
				"-n", namespace,
				"--type", "merge",
				"-p", `{"data":{"image":"tampered-value"}}`)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying the ConfigMap data is restored")
			verifyRestored := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "configmap",
					"ibm-kms-vault-plugin-provider",
					"-n", namespace,
					"-o", "jsonpath={.data.image}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("quay.io/kevinrizza/test-vault-plugin-image:latest"))
			}
			Eventually(verifyRestored).Should(Succeed())
		})
	})
})
