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
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
	"strings"
)

var _ = Describe("Device Discovery Tests", func() {

	var (
		err       error
		k8sClient k8s.K8sClient
	)

	k8sClient, _ = k8s.GetClientSet()
	BeforeEach(func() {
		By("getting a new client set")
		_ = k8sClient.RegenerateClient()
		By("creating the NDM Daemonset")
		err = k8sClient.CreateNDMDaemonSet()
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the daemonset pod to be running")
		ok := WaitForPodToBeRunningEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())

	})
	AfterEach(func() {
		By("deleting the NDM deamonset")
		err := k8sClient.DeleteNDMDaemonSet()
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the pod to be removed")
		ok := WaitForPodToBeDeletedEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	})

	Context("Setup with no external disk", func() {

		It("should have one sparse block device", func() {
			By("regenerating the client set")
			_ = k8sClient.RegenerateClient()

			By("listing all blockdevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			noOfSparseBlockDevices := 0
			// Get the no.of sparse block devices from block device list
			By("counting the number of sparse BDs")
			for _, blockDevice := range bdList.Items {
				if strings.Contains(blockDevice.Name, SparseBlockDeviceName) {
					Expect(blockDevice.Status.State).To(Equal(v1alpha1.BlockDeviceActive))
					noOfSparseBlockDevices++
				}
			}
			Expect(noOfSparseBlockDevices).To(Equal(1))
		})
	})

	Context("Setup with a single external disk already attached", func() {
		var physicalDisk udev.Disk
		By("creating and attaching a new disk")
		physicalDisk = udev.NewDisk(DiskImageSize)
		_ = physicalDisk.AttachDisk()

		It("should have 2 BlockDeviceCR", func() {
			By("regenerating the client set")
			_ = k8sClient.RegenerateClient()

			// should have 2 block device CR, one for sparse disk and one for the
			// external disk
			By("listing all the BlockDevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			noOfSparseBlockDeviceCR := 0
			noOfPhysicalBlockDeviceCR := 0

			// Get no.of sparse blockdevices and physical blockdevices from bdList
			By("counting the number of active sparse and blockdevices")
			for _, blockDevice := range bdList.Items {
				if strings.Contains(blockDevice.Name, BlockDeviceName) && blockDevice.Spec.Path == physicalDisk.Name {
					noOfPhysicalBlockDeviceCR++
				} else if strings.Contains(blockDevice.Name, SparseBlockDeviceName) {
					noOfSparseBlockDeviceCR++
				}
				Expect(blockDevice.Status.State).To(Equal(v1alpha1.BlockDeviceActive))
			}

			Expect(noOfSparseBlockDeviceCR).To(Equal(1))
			Expect(noOfPhysicalBlockDeviceCR).To(Equal(1))
		})
		It("should have blockdeviceCR inactive when disk is detached", func() {
			By("regenerating the client set")
			_ = k8sClient.RegenerateClient()

			By("detaching the disk")
			err = physicalDisk.DetachDisk()
			k8s.WaitForReconciliation()

			By("listing all the blockdevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			noOfPhysicalBlockDeviceCR := 0

			By("checking for inactive blockdevice")
			for _, bd := range bdList.Items {
				if strings.Contains(bd.Name, BlockDeviceName) && bd.Spec.Path == physicalDisk.Name {
					noOfPhysicalBlockDeviceCR++
					Expect(bd.Status.State).To(Equal(v1alpha1.BlockDeviceInactive))
				}
			}

			By("verifying only block device was made inactive")
			Expect(noOfPhysicalBlockDeviceCR).To(Equal(1))
		})
	})
	Context("Setup with a single external disk attached at runtime", func() {
		var physicalDisk udev.Disk
		physicalDisk = udev.NewDisk(DiskImageSize)

		It("should have one additional BlockDevice CR after we attach a disk", func() {
			By("regenerating the client set")
			_ = k8sClient.RegenerateClient()

			By("listing all the blockdevices")
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())
			noOfBlockDeviceCR := len(bdList.Items)

			By("attaching the disk")
			err = physicalDisk.AttachDisk()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconciliation()

			By("listing the blockdevices again")
			bdList, err = k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			By("checking count of blockdevices")
			Expect(len(bdList.Items)).To(Equal(noOfBlockDeviceCR + 1))
		})
	})
})
