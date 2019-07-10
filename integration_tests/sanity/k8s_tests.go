/*
Copyright 2019 The OpenEBS Authors

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

package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	. "github.com/openebs/node-disk-manager/integration_tests/minikube"
)

const (
	// SparseBlockDeviceName is the name given to blockDevice CRs created for a
	// sparse image
	SparseBlockDeviceName = "sparse-"
	// DiskName is the name of the disk CRs created corresponding to
	// physical disks
	DiskName = "disk-"
	// BlockDeviceName is the name of the blockDevice CRs created corresponding to
	// physical/virtual disks or blockdevices
	BlockDeviceName = "blockdevice-"
	// ActiveState stores the active state of Disk/Device resource
	ActiveState = "Active"
	// InactiveState stores the deactivated state of Disk/Device resource
	InactiveState = "Inactive"
	// DiskImageSize is the default file size(1GB) used while creating backing image
	DiskImageSize = 1073741824
)

var _ = Describe("NDM Setup Tests", func() {

	var (
		noOfNodes int
		err       error
	)

	k8sClient, _ := k8s.GetClientSet()
	Context("Checking for Daemonset pods in the cluster", func() {
		BeforeEach(func() {
			nodes, err := k8sClient.ListNodeStatus()
			Expect(err).NotTo(HaveOccurred())
			noOfNodes = len(nodes)
		})
		It("should have running ndm pod on each node after creation", func() {

			err = k8sClient.CreateNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8sClient.ListPodStatus()
			Expect(err).NotTo(HaveOccurred())

			noOfPods := 0
			// Get the number of running pods
			for _, status := range pods {
				if status == Running {
					noOfPods++
				}
			}
			Expect(noOfPods).To(Equal(noOfNodes + 1))
		})
		It("should not have any ndm pods after deletion", func() {
			err = k8sClient.DeleteNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			pods, err := k8sClient.ListPodStatus()
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
