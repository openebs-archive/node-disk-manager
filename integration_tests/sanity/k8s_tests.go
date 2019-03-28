package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	. "github.com/openebs/node-disk-manager/integration_tests/minikube"
)

const (
	SparseDiskName = "sparse-"
	DiskName       = "disk-"
	ActiveState    = "Active"
	InactiveState  = "Inactive"
	DiskImageSize  = 1073741824
)

var _ = Describe("NDM Basic Tests", func() {

	var (
		noOfNodes int
		err       error
	)

	k8sClient, _ := k8s.GetClientSet()
	Context("Checking for Daemonset pods in the cluster", func() {
		BeforeEach(func() {
			nodes, err := k8s.GetNodes(k8sClient.ClientSet)
			Expect(err).NotTo(HaveOccurred())
			noOfNodes = len(nodes)
		})
		It("should have running ndm pod on each node after creation", func() {

			err = k8s.CreateNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8s.GetPods(k8sClient.ClientSet)
			Expect(err).NotTo(HaveOccurred())

			noOfPods := 0
			// Get the number of running pods
			for _, status := range pods {
				if status == Running {
					noOfPods++
				}
			}
			Expect(noOfPods).To(Equal(noOfNodes))
		})
		It("should not have any ndm pods after deletion", func() {
			err = k8s.DeleteNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8s.GetPods(k8sClient.ClientSet)
			Expect(err).NotTo(HaveOccurred())

			noOfPods := 0
			// Get the number of running pods
			for _, status := range pods {
				if status == Running {
					noOfPods++
				}
			}
			Expect(noOfPods).To(BeZero())
		})
	})
})
