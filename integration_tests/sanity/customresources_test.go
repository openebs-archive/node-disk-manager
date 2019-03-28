package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
	"strings"
)

var _ = Describe("Custom Resource Tests", func() {
	var (
		noOfNodes int
		err       error
	)

	k8sClient, _ := k8s.GetClientSet()
	BeforeEach(func() {
		nodes, err := k8s.GetNodes(k8sClient.ClientSet)
		Expect(err).NotTo(HaveOccurred())
		noOfNodes = len(nodes)
	})

	Context("Setup with no external disk", func() {
		BeforeEach(func() {
			err = k8s.CreateNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err := k8s.DeleteNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have one sparse disk per node", func() {
			diskList, err := k8s.GetDiskList(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			noOfDisks := len(diskList.Items)
			Expect(noOfDisks).NotTo(BeZero())

			noOfSparseDisks := 0
			// Get the no.of sparse disks from disk List
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, SparseDiskName) {
					Expect(disk.Status.State).To(Equal(ActiveState))
					noOfSparseDisks++
				}
			}
			Expect(noOfSparseDisks).To(Equal(noOfNodes))
		})
	})

	Context("Setup with a single external disk already attached", func() {
		var physicalDisk udev.Disk
		physicalDisk = udev.NewDisk(DiskImageSize)
		_ = physicalDisk.AttachDisk()
		BeforeEach(func() {
			err = k8s.CreateNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err = k8s.DeleteNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have 2 DiskCRs per node", func() {
			diskList, err := k8s.GetDiskList(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			noOfDisks := len(diskList.Items)
			Expect(noOfDisks).NotTo(BeZero())

			noOfSparseDiskCR := 0
			noOfPhysicalDiskCR := 0
			// Get no.of sparse disk and disk CR from disk List
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, DiskName) && disk.Spec.Path == physicalDisk.Name {
					noOfPhysicalDiskCR++
				} else if strings.Contains(disk.Name, SparseDiskName) {
					noOfSparseDiskCR++
				}
				Expect(disk.Status.State).To(Equal(ActiveState))
			}
			Expect(noOfSparseDiskCR).To(Equal(noOfNodes))
			Expect(noOfPhysicalDiskCR).To(Equal(noOfNodes))
		})
		It("should have physical disk cr inactive when disk is detached", func() {
			err = physicalDisk.DetachDisk()
			k8s.WaitForStateChange()

			diskList, err := k8s.GetDiskList(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			noOfDisks := len(diskList.Items)
			Expect(noOfDisks).NotTo(BeZero())

			noOfPhysicalDiskCR := 0
			// Get no. of disk CRs from disk List
			for _, disk := range diskList.Items {
				if strings.Contains(disk.Name, DiskName) && disk.Spec.Path == physicalDisk.Name {
					noOfPhysicalDiskCR++
					Expect(disk.Status.State).To(Equal(InactiveState))
				}
			}
			Expect(noOfPhysicalDiskCR).To(Equal(noOfNodes))
		})
	})
	Context("Setup with a single external disk attached at runtime", func() {
		var physicalDisk udev.Disk
		physicalDisk = udev.NewDisk(DiskImageSize)
		BeforeEach(func() {
			err = k8s.CreateNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			k8sClient, _ = k8s.GetClientSet()
		})
		AfterEach(func() {
			err = k8s.DeleteNDMYAML(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()
		})
		It("should have one additional disk CR after we attach a disk", func() {
			diskList, err := k8s.GetDiskList(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			noOfDiskCR := len(diskList.Items)
			Expect(noOfDiskCR).NotTo(BeZero())

			err = physicalDisk.AttachDisk()
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			diskList, err = k8s.GetDiskList(k8sClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(diskList.Items)).To(Equal(noOfDiskCR + 1))
		})
	})
})
