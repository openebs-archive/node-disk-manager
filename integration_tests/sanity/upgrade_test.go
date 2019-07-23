package sanity

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
)

const (
	// oldBDCFinalizer is the old string from which BDC should be updated
	oldBDCFinalizer = "blockdeviceclaim.finalizer"
	// newBDCFinalizer is the new string to which BDC to be updated
	newBDCFinalizer = "openebs.io/bdc-protection"
	//
)

var _ = Describe("Pre upgrade tests", func() {

	k8sClient, _ := k8s.GetClientSet()

	BeforeEach(func() {
		err := k8sClient.CreateNDMYAML()
		Expect(err).NotTo(HaveOccurred())
		k8s.WaitForStateChange()
		k8sClient, _ = k8s.GetClientSet()
	})
	AfterEach(func() {
		err := k8sClient.DeleteNDMYAML()
		Expect(err).NotTo(HaveOccurred())
		k8s.WaitForStateChange()
	})
	Context("BDC with old finalizer", func() {
		bdcName1 := "test-bdc1"
		var blockDeviceClaim *apis.BlockDeviceClaim
		BeforeEach(func() {
			blockDeviceClaim = k8s.NewBDC(bdcName1)
		})
		It("has only the old BDC finalizer", func() {
			blockDeviceClaim.Finalizers = append(blockDeviceClaim.ObjectMeta.Finalizers, oldBDCFinalizer)
			blockDeviceClaim.Spec.BlockDeviceName = FakeBlockDevice

			// create the BDC with old finalizer
			err := k8sClient.CreateBlockDeviceClaim(blockDeviceClaim)
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForReconcilation()

			// restart the ndm operator pod
			err = k8sClient.RestartPod("node-disk-operator")
			Expect(err).NotTo(HaveOccurred())
			k8s.WaitForStateChange()

			// list BDC and check for new finalizer
			bdcList, err := k8sClient.ListBlockDeviceClaims()
			Expect(err).NotTo(HaveOccurred())
			Expect(bdcList.Items).To(Equal(1))

			for _, bdc := range bdcList.Items {
				Expect(bdc.Finalizers).To(Equal(newBDCFinalizer))
			}

		})

	})
})
