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
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog"

	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/udev"
)

var _ = FDescribe("Mount-point and fs type change detection tests", func() {
	var err error
	var k8sClient k8s.K8sClient
	physicalDisk := udev.NewDisk(DiskImageSize)

	mountpath, err := ioutil.TempDir("", "ndm-integration-tests")
	if err != nil {
		Fail("setup failed")
	}
	err = physicalDisk.CreateFileSystem()
	if err != nil {
		Fail("setup failed")
	}

	k8sClient, err = k8s.GetClientSet()
	if err != nil {
		Fail("setup failed")
	}

	It("should create ndm daemonset", func() {
		err = k8sClient.CreateNDMDaemonSet()
		Expect(err).ToNot(HaveOccurred())

		ok := WaitForPodToBeRunningEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	})

	Context("Device attached", func() {
		var bdName, bdNamespace string

		It("should attach disk", func() {
			err := physicalDisk.AttachDisk()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should be added to etcd", func() {
			found := false
			bdlist, err := k8sClient.ListBlockDevices()
			Expect(err).ToNot(HaveOccurred())
			Expect(bdlist).ToNot(BeNil())
			klog.Info("want ", physicalDisk.Name)
			for _, bd := range bdlist.Items {
				klog.Info(bd.Name, bd.Spec.Path)
				if bd.Spec.Path == physicalDisk.Name {
					found = true
					bdName = bd.Name
					bdNamespace = bd.Namespace
				}
			}
			Expect(found).To(BeTrue())
		})

		Context("mounting device", func() {
			It("device should mount", func() {
				Expect(physicalDisk.Mount(mountpath)).ToNot(HaveOccurred())
				k8s.WaitForStateChange()
			})

			It("should add mount-point to blockdevice", func() {
				bd, err := k8sClient.GetBlockDevice(bdName, bdNamespace)
				Expect(err).ToNot(HaveOccurred())
				Expect(bd).ToNot(BeNil())
				Expect(bd.Name).To(Equal(bdName))
				Expect(bd.Namespace).To(Equal(bdNamespace))
				Expect(bd.Spec.FileSystem.Mountpoint).To(Equal(mountpath))
			})
		})

		Context("unmounting device", func() {
			It("should unmount", func() {
				Expect(physicalDisk.Unmount()).ToNot(HaveOccurred())
				k8s.WaitForStateChange()
			})

			It("should remove mount-point from blockdevice", func() {
				bd, err := k8sClient.GetBlockDevice(bdName, bdNamespace)
				Expect(err).ToNot(HaveOccurred())
				Expect(bd).ToNot(BeNil())
				Expect(bd.Name).To(Equal(bdName))
				Expect(bd.Namespace).To(Equal(bdNamespace))
				Expect(bd.Spec.FileSystem.Mountpoint).To(BeEmpty())
			})

			It("should detach and delete disk", func() {
				Expect(physicalDisk.DetachAndDeleteDisk()).ToNot(HaveOccurred())
			})
		})
	})

	It("should delete ndm daemonset", func() {
		k8s.WaitForReconciliation()

		err := k8sClient.DeleteNDMDaemonSet()
		Expect(err).ToNot(HaveOccurred())

		ok := WaitForPodToBeDeletedEventually(DaemonSetPodPrefix)
		Expect(ok).To(BeTrue())
	})

})
