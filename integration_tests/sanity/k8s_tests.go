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
	"strings"
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
	// DiskImageSize is the default file size(1GB) used while creating backing image
	DiskImageSize = 1073741824
	// OperatorPodPrefix is the pod name for NDM operator
	OperatorPodPrefix = "node-disk-operator"
	// DaemonSetPodPrefix is the pod name for NDM daemon
	DaemonSetPodPrefix = "node-disk-manager"
)

var _ = Describe("NDM Setup Tests", func() {

	var err error

	k8sClient, _ := k8s.GetClientSet()
	Context("Checking for Daemonset pods in the cluster", func() {

		It("should have running ndm pod on each node after installation", func() {

			By("creating NDM daemonset")
			err = k8sClient.CreateNDMDaemonSet()
			Expect(err).NotTo(HaveOccurred())

			By("waiting for daemonset pods to be in running state")
			ok := WaitForPodToBeRunningEventually(DaemonSetPodPrefix)
			Expect(ok).To(BeTrue())
		})

		It("should not have any ndm pods after deletion", func() {

			By("deleting NDM deamonset")
			err = k8sClient.DeleteNDMDaemonSet()
			Expect(err).NotTo(HaveOccurred())

			By("no of daemonset pods should be zero")
			ok := WaitForPodToBeDeletedEventually(DaemonSetPodPrefix)
			Expect(ok).To(BeTrue())
		})
	})
})

// WaitForPodToBeRunningEventually waits for 2 minutes for the given pod to be
// in running state
func WaitForPodToBeRunningEventually(podPrefix string) bool {
	return Eventually(func() string {
		c, _ := k8s.GetClientSet()
		pods, err := c.ListPodStatus()
		Expect(err).NotTo(HaveOccurred())
		for pod, state := range pods {
			if strings.Contains(pod, podPrefix) {
				return state
			}
		}
		return ""
	}, 120, 5).Should(Equal(k8s.Running))
}

// WaitForPodToBeDeletedEventually waits for 2 minutes for the given pod to
// get deleted
func WaitForPodToBeDeletedEventually(podPrefix string) bool {
	return Eventually(func() int {
		c, _ := k8s.GetClientSet()
		pods, err := c.ListPodStatus()
		Expect(err).NotTo(HaveOccurred())

		noOfPods := 0
		for pod := range pods {
			if strings.Contains(pod, podPrefix) {
				noOfPods++
			}
		}
		return noOfPods
	}, 120, 5).Should(BeZero())
}
