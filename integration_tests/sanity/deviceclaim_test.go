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
	"github.com/openebs/node-disk-manager/integration_tests/udev"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
)

const (
	// FakeHostName is a generated fake hostname
	FakeHostName = "fake-minikube"
	// FakeBlockDevice is a generated fake block device name
	FakeBlockDevice = "fake-BD"
)

var (
	BDCUnavailableCapacity = resource.MustParse("10Gi")
	BDCAvailableCapacity   = resource.MustParse("1Gi")
	// HostName is the hostname in which the tests are performed
	HostName = os.Getenv("HOSTNAME")
)

var _ = Describe("BlockDevice Claim tests", func() {

	var err error
	var k8sClient k8s.K8sClient
	physicalDisk := udev.NewDisk(DiskImageSize)
	_ = physicalDisk.AttachDisk()

	BeforeEach(func() {
		By("getting a new client set")
		k8sClient, err = k8s.GetClientSet()
		Expect(err).NotTo(HaveOccurred())

		By("creating the NDM Daemonset")
		err = k8sClient.CreateNDMDaemonSet()
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the daemonset pod to be running")
		ok := WaitForPodToBeRunningEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())

		k8s.WaitForReconciliation()
	})
	AfterEach(func() {
		By("deleting the NDM deamonset")
		err := k8sClient.DeleteNDMDaemonSet()
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the pod to be removed")
		ok := WaitForPodToBeDeletedEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	})
	Context("Claim Block Device when matching BD is not available", func() {
		var bdcName string
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			By("building a new BDC")
			bdcName = "test-bdc-1"
			blockDeviceClaim = k8s.NewBDC(bdcName)
		})
		AfterEach(func() {
			// delete the BDC
			By("deleting the BDC as part of cleanup")
			err = k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
		})
		It("BD is not available on the host", func() {
			blockDeviceClaim.Spec.HostName = FakeHostName
			blockDeviceClaim.Namespace = k8s.Namespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity

			By("creating BDC object")
			err = k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			By("listing all BDCs")
			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			By("check whether BDC is in Pending state")
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusPending))
				}
			}
		})

		It("BD with resource requirement is not available on the host", func() {
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = k8s.Namespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCUnavailableCapacity

			By("creating BDC object with unavailable capacity")
			err = k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			By("listing all BDCs")
			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			By("check whether BDC is in Pending state")
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusPending))
				}
			}
		})
	})

	Context("Claim Block Device when matching BD is available", func() {
		var bdcName string
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			By("building a new BDC")
			bdcName = "test-bdc-1"
			blockDeviceClaim = k8s.NewBDC(bdcName)
		})
		AfterEach(func() {
			By("getting the BDC from etcd")
			blockDeviceClaim, err = k8sClient.GetBlockDeviceClaim(blockDeviceClaim.Namespace, blockDeviceClaim.Name)
			Expect(err).NotTo(HaveOccurred())

			By("remove finalizer on the BDC")
			blockDeviceClaim.Finalizers = nil
			err = k8sClient.UpdateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())

			// delete the BDC
			By("deleting the BDC as part of cleanup")
			err = k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())

		})
		It("has matching BD on the node", func() {
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = k8s.Namespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity

			By("creating BDC with matching node")
			err = k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			By("listing all BDCs")
			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			var bdName string
			// check status of BDC
			By("checking if BDC is bound")
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					bdName = bdc.Spec.BlockDeviceName
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusDone))
				}
			}

			// check status of BD that has been bound
			By("listing all blockdevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			By("checking if the corresponding BD is claimed")
			for _, bd := range bdList.Items {
				if bd.Name == bdName {
					Expect(bd.Status.ClaimState).To(Equal(apis.BlockDeviceClaimed))
				}
			}

		})
	})
	Context("Unclaiming a block device ", func() {
		var bdcName string
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			By("building a new BDC")
			bdcName = "test-bdc-1"
			blockDeviceClaim = k8s.NewBDC(bdcName)
		})
		It("unclaims a BD when BDC is deleted", func() {
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = k8s.Namespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity

			By("creating BDC object")
			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			By("listing all BDCs")
			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			var bdName string
			// check status of BDC
			By("checking if BDC is bound")
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					bdName = bdc.Spec.BlockDeviceName
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusDone))
				}
			}

			// check status of BD that has been bound
			By("listing all blockdevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			By("checking if the corresponding BD is claimed")
			for _, bd := range bdList.Items {
				if bd.Name == bdName {
					Expect(bd.Status.ClaimState).To(Equal(apis.BlockDeviceClaimed))
				}
			}

			By("deleting the BDC")
			err = k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			// check status of BD again to check it has been released
			bdList, err = k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())
			By("checking if the corresponding BD is in released/unclaimed state")
			for _, bd := range bdList.Items {
				if bd.Name == bdName {
					// BlockDevice can be in either released or unclaimed
					// depending on the time required for cleanup
					Expect(bd.Status.ClaimState).To(Or(Equal(apis.BlockDeviceReleased),
						Equal(apis.BlockDeviceUnclaimed)))
				}
			}

		})
	})

})
