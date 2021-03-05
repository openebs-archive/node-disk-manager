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
	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
)

const (
	// oldBDCFinalizer is the old string from which BDC should be updated
	oldBDCFinalizer = "blockdeviceclaim.finalizer"
	// newBDCFinalizer is the new string to which BDC to be updated
	newBDCFinalizer = "openebs.io/bdc-protection"
)

var _ = Describe("[upgrade] TEST PRE-UPGRADES IN NDM OPERATOR", func() {
	var err error
	k8sClient, _ := k8s.GetClientSet()

	Context("[finalizer] [rename] BDC with old finalizer", func() {
		bdcName1 := "test-bdc1"
		var blockDeviceClaim *apis.BlockDeviceClaim

		BeforeEach(func() {
			By("building BDC object")
			blockDeviceClaim = k8s.NewBDC(bdcName1)
			_ = k8sClient.RegenerateClient()
		})

		AfterEach(func() {
			// get the BDC
			By("cleaning up BDC after test", func() {
				By("getting the BDC from etcd")
				blockDeviceClaim, err = k8sClient.GetBlockDeviceClaim(blockDeviceClaim.Namespace, blockDeviceClaim.Name)
				Expect(err).NotTo(HaveOccurred())

				// remove finalizer on BDC so that it can be deleted cleanly
				By("remove finalizer on the BDC")
				blockDeviceClaim.Finalizers = nil
				err = k8sClient.UpdateBlockDeviceClaim(blockDeviceClaim)
				Expect(err).NotTo(HaveOccurred())

				// delete the BDC
				By("deleting the BDC as part of cleanup")
				err = k8sClient.DeleteBlockDeviceClaim(blockDeviceClaim)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("has only the old BDC finalizer", func() {
			By("adding the old finalizer to BDC")
			blockDeviceClaim.Finalizers = append(blockDeviceClaim.ObjectMeta.Finalizers, oldBDCFinalizer)
			blockDeviceClaim.Namespace = k8s.Namespace
			blockDeviceClaim.Spec.BlockDeviceName = FakeBlockDevice

			// create the BDC with old finalizer
			By("creating the BDC object in etcd")
			err = k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())

			// restart the ndm operator pod
			By("restarting the NDM operator pod to simulate a fresh start")
			err = k8sClient.RestartPod(OperatorPodPrefix)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for pod to be in running state")
			ok := WaitForPodToBeRunningEventually(OperatorPodPrefix)
			Expect(ok).To(BeTrue())

			// list BDC and check for new finalizer
			By("checking for the new finalizer in the BDC")
			Eventually(func() []string {
				blockDeviceClaim, err = k8sClient.GetBlockDeviceClaim(blockDeviceClaim.Namespace, blockDeviceClaim.Name)
				Expect(err).NotTo(HaveOccurred())
				return blockDeviceClaim.Finalizers
			}, 120, 5).Should(ContainElement(newBDCFinalizer))

		})

	})
})
