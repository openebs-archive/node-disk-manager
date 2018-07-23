package integrationtest

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/CITF/utils/k8s"
	cr "github.com/openebs/node-disk-manager/integration_test/common_resource"
	"github.com/openebs/node-disk-manager/integration_test/ndm_util"
)

func TestIntegrationNDM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {
	var err error

	// It starts minikube if it is not Running
	cr.CitfInstance.Environment.Setup()

	// if you are using minikube version greater than 0.24.1
	// then you have to update the K8s config
	// this extra step will be unsolicited in upcoming changes.
	cr.CitfInstance.K8S, err = k8s.NewK8S()
	Expect(err).NotTo(HaveOccurred())

	// It waits till namespace is ready
	ndmutil.WaitTillDefaultNSisReady()

	// It prepares configuration and Applies the same
	ndmutil.ReplaceImageInYAMLAndApply()

	// It waits till node-disk-manager is ready or timeout reached
	err = ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	// It Delete minikube if it is running
	// It stops residue containers
	// It removes remaining residue files
	ndmutil.Clean()
})

var _ = Describe("Integration Test", func() {
	// Now as BeforeSuite has run, We shall have a healthy node-disk-manager daemonset
	When("We check the log", func() {
		It("has `started the controller` in the log", func() {
			validated, err := ndmutil.GetNDMLogAndValidate()

			Expect(err).NotTo(HaveOccurred())
			Expect(validated).To(BeTrue())
		})
	})

	When("We check the disks", func() {
		Specify("`ndm device list` output inside the node-disk-manager pod "+
			"and `lsblk -brdno name,size,model` output on the host should match", func() {
			matched, err := ndmutil.MatchDisksOutsideAndInside()

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})
})
