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
	"github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"strings"
)

var _ = Describe("Device Discovery Tests", func() {
	var (
		noOfNodes int
		err       error
	)

	k8sClient, _ := k8s.GetClientSet()
	BeforeEach(func() {
		nodes, err := k8sClient.ListNodeStatus()
		Expect(err).NotTo(HaveOccurred())
		noOfNodes = len(nodes)
	})

	Context("Setup with no external disk", func() {
		BeforeEach(func() {
			err = k8sClient.CreateNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err := k8sClient.DeleteNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have one sparse block device per node", func() {
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			noOfSparseBlockDevices := 0
			// Get the no.of sparse block devices from block device list
			for _, blockDevice := range bdList.Items {
				if strings.Contains(blockDevice.Name, SparseBlockDeviceName) {
					Expect(blockDevice.Status.State).To(Equal(v1alpha1.BlockDeviceActive))
					noOfSparseBlockDevices++
				}
			}
			Expect(noOfSparseBlockDevices).To(Equal(noOfNodes))
		})
	})

	Context("Setup with a single external disk already attached", func() {
		var physicalDisk udev.Disk
		physicalDisk = udev.NewDisk(DiskImageSize)
		_ = physicalDisk.AttachDisk()
		BeforeEach(func() {
			err = k8sClient.CreateNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err = k8sClient.DeleteNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have 1 DiskCR and 2 BlockDeviceCR per node", func() {
			// should have 2 block device CR, one for sparse disk and one for the
			// external disk
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			// should have single DiskCR which corresponds to the external disk attached
			diskList, err := k8sClient.ListDisk()
			Expect(err).NotTo(HaveOccurred())

			noOfPhysicalDiskCR := 0
			noOfSparseBlockDeviceCR := 0
			noOfPhysicalBlockDeviceCR := 0

			// Get no.of sparse blockdevices and physical blockdevices from bdList
			for _, blockDevice := range bdList.Items {
				if strings.Contains(blockDevice.Name, BlockDeviceName) && blockDevice.Spec.Path == physicalDisk.Name {
					noOfPhysicalBlockDeviceCR++
				} else if strings.Contains(blockDevice.Name, SparseBlockDeviceName) {
					noOfSparseBlockDeviceCR++
				}
				Expect(blockDevice.Status.State).To(Equal(v1alpha1.BlockDeviceActive))
			}

			// Get no of physical disk CRs from diskList
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, DiskName) && disk.Spec.Path == physicalDisk.Name {
					noOfPhysicalDiskCR++
				}
				Expect(disk.Status.State).To(Equal(v1alpha1.BlockDeviceActive))
			}

			Expect(noOfPhysicalDiskCR).To(Equal(1))
			Expect(noOfSparseBlockDeviceCR).To(Equal(noOfNodes))
			Expect(noOfPhysicalBlockDeviceCR).To(Equal(1))
		})
		It("should have diskCR && blockdeviceCR inactive when disk is detached", func() {
			err = physicalDisk.DetachDisk()
			k8s.WaitForStateChange()

			diskList, err := k8sClient.ListDisk()
			Expect(err).NotTo(HaveOccurred())

			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			// the disk CR should be inactive
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, DiskName) && disk.Spec.Path == physicalDisk.Name {
					Expect(disk.Status.State).To(Equal(v1alpha1.DiskInactive))
				}
			}

			for _, bd := range bdList.Items {
				if strings.Contains(bd.Name, BlockDeviceName) && bd.Spec.Path == physicalDisk.Name {
					Expect(bd.Status.State).To(Equal(v1alpha1.DiskInactive))
				}
			}
		})
	})
	Context("Setup with a single external disk attached at runtime", func() {
		var physicalDisk udev.Disk
		physicalDisk = udev.NewDisk(DiskImageSize)
		BeforeEach(func() {
			err = k8sClient.CreateNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err = k8sClient.DeleteNDMYAML()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have one additional Disk and BlockDevice CR after we attach a disk", func() {
			diskList, err := k8sClient.ListDisk()
			Expect(err).NotTo(HaveOccurred())
			noOfDiskCR := len(diskList.Items)

			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())
			noOfBlockDeviceCR := len(bdList.Items)

			err = physicalDisk.AttachDisk()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			diskList, err = k8sClient.ListDisk()
			Expect(err).NotTo(HaveOccurred())
			bdList, err = k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(diskList.Items)).To(Equal(noOfDiskCR + 1))
			Expect(len(bdList.Items)).To(Equal(noOfBlockDeviceCR + 1))
		})
	})
})
