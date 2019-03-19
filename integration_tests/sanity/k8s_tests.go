package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	. "github.com/openebs/node-disk-manager/integration_tests/minikube"
)

const (
	namespace = "default"
)

var _ = Describe("NDM Basic Tests", func() {

	clientSet, _ := k8s.GetClientSet()
	Context("Checking for pods in the cluster", func() {
		It("should not have any pods", func() {
			pods, err := k8s.GetPods(clientSet, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pods)).To(BeZero())
		})
	})

	Context("Applying the ndm yaml", func() {
		It("should have running Daemonset pod on each node", func() {

			nodes, err := k8s.GetNodes(clientSet)
			Expect(err).NotTo(HaveOccurred())
			noOfNodes := len(nodes)

			err = k8s.CreateNDMYAML(clientSet, namespace)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8s.GetPods(clientSet, namespace)
			Expect(err).NotTo(HaveOccurred())
			noOfPods := len(pods)

			Expect(noOfPods).To(Equal(noOfNodes))

			for _, status := range pods {
				Expect(status).To(Equal(Running))
			}
		})
	})

	Context("Deleting the ndm yaml", func() {
		It("should not have any ndm pods", func() {
			err := k8s.DeleteNDMYAML(clientSet, namespace)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8s.GetPods(clientSet, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pods)).To(BeZero())
		})
	})

})
