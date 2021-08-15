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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
)

const newDiskCapacity int64 = 500 * 1024 * 1024

var _ = Describe("Size change detection tests", func() {
	var kcli k8s.K8sClient
	disk := udev.NewDisk(DiskImageSize)

	var bdName, bdNamespace string

	BeforeEach(func() {
		By("initializing up k8s cli", initK8sCli(&kcli))
		By("creating up ndm daemonset", createAndStartNDMDaemonset(&kcli))
		By("attaching disk", attachDisk(&disk))
		By("waiting for etcd update", k8s.WaitForStateChange)
		By("verifying disk added to etcd", verifyDiskAddedToEtcd(&kcli, disk.Name,
			&bdName, &bdNamespace))
	})
	AfterEach(func() {
		By("detaching disk", detachDisk(&disk))
		By("destroying ndm daemonset", stopAndDeleteNDMDaemonset(&kcli))
	})

	It("should detect size change and update bd capacity", func() {
		By("Changing disk size", changeDiskSize(&disk))
		By("Waiting for etcd update", k8s.WaitForStateChange)
		By("Verifying bd capacity update in etcd",
			verifyDiskCapacityUpdate(&kcli, bdName, bdNamespace,
				Equal(uint64(newDiskCapacity))))

	})
})

func changeDiskSize(disk *udev.Disk) func() {
	return func() {
		Expect(disk.Resize(newDiskCapacity)).ToNot(HaveOccurred())
	}
}

func verifyDiskCapacityUpdate(cli *k8s.K8sClient, bdName, bdNamespace string,
	matcher types.GomegaMatcher) func() {
	return func() {
		bd, err := cli.GetBlockDevice(bdName, bdNamespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(bd).ToNot(BeNil())
		Expect(bd.Name).To(Equal(bdName))
		Expect(bd.Namespace).To(Equal(bdNamespace))
		Expect(bd.Spec.Capacity.Storage).To(matcher)
	}
}
