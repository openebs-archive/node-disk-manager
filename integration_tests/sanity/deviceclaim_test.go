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
)

const (
	// DefaultNamespace is the default namespace in a k8s cluster
	DefaultNamespace = "default"
	// HostName is the hostname in which the tests are performed
	HostName = "minikube"
	// FakeHostName is a generated fake hostname
	FakeHostName = "fake-minikube"
)

var (
	BDCUnavailableCapacity = resource.MustParse("10Gi")
	BDCAvailableCapacity   = resource.MustParse("1Gi")
	BDCInvalidCapacity     = resource.MustParse("0")
)

var _ = Describe("BlockDevice Claim tests", func() {

	// attach a new physical disk
	physicalDisk := udev.NewDisk(DiskImageSize)
	_ = physicalDisk.AttachDisk()

	k8sClient, _ := k8s.GetClientSet()

	BeforeEach(func() {
		err := k8sClient.CreateNDMYAML()
		Expect(err).NotTo(HaveOccurred())
		k8s.WaitForStateChange()
		k8sClient, _ = k8s.GetClientSet()
		//k8sClient.WaitForBDC()
	})
	AfterEach(func() {
		err := k8sClient.DeleteNDMYAML()
		Expect(err).NotTo(HaveOccurred())
		k8s.WaitForStateChange()
	})

	Context("Claim Block Device when matching BD is not available", func() {
		bdcName := "test-blockdeviceclaim"
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			blockDeviceClaim = k8s.NewBDC(bdcName)
		})
		AfterEach(func() {
			err := k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
		})
		It("BD is not available on the host", func() {
			blockDeviceClaim.Spec.HostName = FakeHostName
			blockDeviceClaim.Namespace = DefaultNamespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity

			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusPending))
				}
			}
		})
		It("BD with resource requirement is not available on the host", func() {
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = DefaultNamespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCUnavailableCapacity

			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusPending))
				}
			}
		})
	})

	Context("Claim Block Device when matching BD is available", func() {
		bdcName := "test-blockdeviceclaim"
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			blockDeviceClaim = k8s.NewBDC(bdcName)
		})
		AfterEach(func() {
			err := k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
		})
		It("has matching BD on the node", func() {
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = DefaultNamespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity

			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			var bdName string
			// check status of BDC
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					bdName = bdc.Spec.BlockDeviceName
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusDone))
				}
			}

			// check status of BD that has been bound
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			for _, bd := range bdList.Items {
				if bd.Name == bdName {
					Expect(bd.Status.ClaimState).To(Equal(apis.BlockDeviceClaimed))
				}
			}

		})
	})
	Context("Unclamining a block device ", func() {
		It("unclaimes a BD when BDC is deleted", func() {
			bdcName := "test-blockdeviceclaim"
			blockDeviceClaim := k8s.NewBDC(bdcName)
			blockDeviceClaim.Spec.HostName = HostName
			blockDeviceClaim.Namespace = DefaultNamespace
			blockDeviceClaim.Spec.Resources.Requests[apis.ResourceStorage] = BDCAvailableCapacity
			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(bdcList.Items)).NotTo(BeZero())

			var bdName string
			// check status of BDC
			for _, bdc := range bdcList.Items {
				if bdc.Name == bdcName {
					bdName = bdc.Spec.BlockDeviceName
					Expect(bdc.Status.Phase).To(Equal(apis.BlockDeviceClaimStatusDone))
				}
			}

			// check status of BD that has been bound
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			for _, bd := range bdList.Items {
				if bd.Name == bdName {
					Expect(bd.Status.ClaimState).To(Equal(apis.BlockDeviceClaimed))
				}
			}

			err = k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			// check status of BD again to check it has been released
			bdList, err = k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())
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
