package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	. "github.com/openebs/node-disk-manager/integration_tests/minikube"
	"testing"
)

var (
	minikube = NewMinikube()
)

func TestNDM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

// Initialize the minikube cluster
var _ = BeforeSuite(func() {
	var err error
	Expect(minikube.IsUpAndRunning()).To(BeTrue())
	err = minikube.WaitForMinikubeToBeReady()
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Verify Kubernetes Cluster Setup", func() {
	var err error
	Context("Initially, we check minikube status", func() {
		It("should be running", func() {
			Expect(minikube.IsUpAndRunning()).To(BeTrue())
		})
	})
	Context("We check for generated Kube Config", func() {
		_, err = k8s.GetClientSet()
		It("should be able to generate ClientSet", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
