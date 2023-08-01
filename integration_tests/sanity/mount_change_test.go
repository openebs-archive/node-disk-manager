/*
Copyright 2020 The OpenEBS Authors

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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
)

var _ = Describe("Mount-point and fs type change detection tests", func() {
	var k8sClient k8s.K8sClient
	physicalDisk := udev.NewDisk(DiskImageSize)

	var bdName, bdNamespace, mountPath string

	BeforeEach(setUp(&k8sClient, &physicalDisk, &bdName, &bdNamespace, &mountPath))
	AfterEach(tearDown(&k8sClient, &physicalDisk))

	It("should pass mount-umount flow", func() {
		// Mount flow
		By("Running mount flow")
		By("Mounting disk", mountDisk(&physicalDisk, mountPath))
		By("Waiting for etcd update", k8s.WaitForStateChange)
		By("Verifying bd mount point update in etcd",
			verifyMountPointUpdate(&k8sClient, bdName, bdNamespace,
				Equal(mountPath)))

		// Umount flow
		By("Running umount flow")
		By("Umounting disk", umountDisk(&physicalDisk))
		By("Waiting for etcd update", k8s.WaitForStateChange)
		By("Verifying bd mount point update in etcd",
			verifyMountPointUpdate(&k8sClient, bdName, bdNamespace,
				BeEmpty()))
	})
})

func setUp(kcli *k8s.K8sClient, disk *udev.Disk, bdName, bdNamespace, mountPath *string) func() {
	return func() {
		By("initializing up k8s cli", initK8sCli(kcli))
		By("creating up ndm daemonset", createAndStartNDMDaemonset(kcli))
		By("setting up fs on disk", setupPhysicalDisk(disk))
		By("attaching disk", attachDisk(disk))
		By("generating random mount path", generateMountPath(mountPath))
		By("waiting for etcd update", k8s.WaitForStateChange)
		By("verifying disk added to etcd", verifyDiskAddedToEtcd(kcli, disk.Name,
			bdName, bdNamespace))
	}
}

func tearDown(kcli *k8s.K8sClient, disk *udev.Disk) func() {
	return func() {
		By("detaching disk", detachDisk(disk))
		By("destroying ndm daemonset", stopAndDeleteNDMDaemonset(kcli))
	}

}

func generateMountPath(mountPath *string) func() {
	return func() {
		mp, err := os.MkdirTemp("", "ndm-integration-tests")
		Expect(err).ToNot(HaveOccurred())
		*mountPath = mp
	}
}

func initK8sCli(cli *k8s.K8sClient) func() {
	return func() {
		kcli, err := k8s.GetClientSet()
		Expect(err).ToNot(HaveOccurred())
		*cli = kcli
	}
}

func setupPhysicalDisk(disk *udev.Disk) func() {
	return func() {
		Expect(disk.CreateFileSystem()).ToNot(HaveOccurred())
	}
}

func createAndStartNDMDaemonset(cli *k8s.K8sClient) func() {
	return func() {
		err := cli.CreateNDMDaemonSet()
		Expect(err).ToNot(HaveOccurred())

		ok := WaitForPodToBeRunningEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	}
}

func stopAndDeleteNDMDaemonset(cli *k8s.K8sClient) func() {
	return func() {
		k8s.WaitForReconciliation()

		err := cli.DeleteNDMDaemonSet()
		Expect(err).ToNot(HaveOccurred())

		ok := WaitForPodToBeDeletedEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	}
}

func verifyMountPointUpdate(cli *k8s.K8sClient, bdName, bdNamespace string,
	matcher types.GomegaMatcher) func() {
	return func() {
		bd, err := cli.GetBlockDevice(bdName, bdNamespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(bd).ToNot(BeNil())
		Expect(bd.Name).To(Equal(bdName))
		Expect(bd.Namespace).To(Equal(bdNamespace))
		Expect(bd.Spec.FileSystem.Mountpoint).To(matcher)
	}
}

func mountDisk(disk *udev.Disk, mountPath string) func() {
	return func() {
		Expect(disk.Mount(mountPath)).ToNot(HaveOccurred())
	}
}

func umountDisk(disk *udev.Disk) func() {
	return func() {
		Expect(disk.Unmount()).ToNot(HaveOccurred())
	}
}

func attachDisk(disk *udev.Disk) func() {
	return func() {
		Expect(disk.AttachDisk()).ToNot(HaveOccurred())
	}
}

func detachDisk(disk *udev.Disk) func() {
	return func() {
		Expect(disk.DetachAndDeleteDisk()).ToNot(HaveOccurred())
	}
}

func verifyDiskAddedToEtcd(cli *k8s.K8sClient, diskName string, bdName *string, bdNamespace *string) func() {
	return func() {
		found := false
		bdlist, err := cli.ListBlockDevices()
		Expect(err).ToNot(HaveOccurred())
		Expect(bdlist).ToNot(BeNil())
		for _, bd := range bdlist.Items {
			if bd.Spec.Path == diskName {
				found = true
				*bdName = bd.Name
				*bdNamespace = bd.Namespace
			}
		}
		Expect(found).To(BeTrue())
	}
}
