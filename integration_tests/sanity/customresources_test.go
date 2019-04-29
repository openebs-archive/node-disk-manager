package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
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
<<<<<<< HEAD
		It("should have one sparse block device per node", func() {
			bdList, err := k8sClient.ListBlockDevices()
=======
		It("should have one sparse disk per node", func() {
			diskList, err := k8sClient.ListDisk()
>>>>>>> refactor(code): address review comments
			Expect(err).NotTo(HaveOccurred())

			noOfSparseBlockDevices := 0
			// Get the no.of sparse block devices from block device list
			for _, blockDevice := range bdList.Items {
				if strings.Contains(blockDevice.Name, SparseBlockDeviceName) {
					Expect(blockDevice.Status.State).To(Equal(ActiveState))
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
<<<<<<< HEAD
		It("should have 1 DiskCR and 2 BlockDeviceCR per node", func() {
			// should have 2 block device CR, one for sparse disk and one for the
			// external disk
			bdList, err := k8sClient.ListBlockDevices()
			Expect(err).NotTo(HaveOccurred())

			// should have single DiskCR which corresponds to the external disk attached
=======
		It("should have 2 DiskCRs per node", func() {
>>>>>>> refactor(code): address review comments
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
				Expect(blockDevice.Status.State).To(Equal(ActiveState))
			}

			// Get no of physical disk CRs from diskList
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, DiskName) && disk.Spec.Path == physicalDisk.Name {
					noOfPhysicalDiskCR++
				}
				Expect(disk.Status.State).To(Equal(ActiveState))
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
					Expect(disk.Status.State).To(Equal(InactiveState))
				}
			}

			for _, bd := range bdList.Items {
				if strings.Contains(bd.Name, BlockDeviceName) && bd.Spec.Path == physicalDisk.Name {
					Expect(bd.Status.State).To(Equal(InactiveState))
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
<<<<<<< HEAD
		It("should have one additional Disk and BlockDevice CR after we attach a disk", func() {
=======
		It("should have one additional disk CR after we attach a disk", func() {
>>>>>>> refactor(code): address review comments
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
