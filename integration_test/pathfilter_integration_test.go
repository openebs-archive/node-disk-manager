package integrationtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_test/ndm_util"
	"time"
)

var _ = Describe("Path filter integration Test", func() {

	var configMap ndmutil.ConfigMap
	var err error

	When("Path filter is disabled", func() {
		err = ndmutil.InitMinikube()
		Expect(err).NotTo(HaveOccurred())

		configMap = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
		configMap.SetPathFilter("false")

		ndmutil.ReplaceAndApplyConfig(configMap)
		// It waits till node-disk-manager is ready or timeout reached
		err = ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		It("has matching disks inside node and the host", func() {

			matched, err := ndmutil.MatchDisksOutsideAndInside()

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})

	When("A device is included", func() {
		err = ndmutil.InitMinikube()
		Expect(err).NotTo(HaveOccurred())

		configMap = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
		configMap.SetIncludePath("/dev/sda")
		ndmutil.ReplaceAndApplyConfig(configMap)
		// It waits till node-disk-manager is ready or timeout reached
		err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		It("has only /dev/sda in the pod", func() {
			matched, err := ndmutil.MatchNDMDeviceList(false, "/dev/sda")

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})

	When("A device is excluded", func() {
		err = ndmutil.InitMinikube()
		Expect(err).NotTo(HaveOccurred())

		configMap = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
		configMap.SetExcludePath("/dev/sda")
		ndmutil.ReplaceAndApplyConfig(configMap)
		// It waits till node-disk-manager is ready or timeout reached
		err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		It("does not have /dev/sda in the pod", func() {
			matched, err := ndmutil.MatchNDMDeviceList(true, "/dev/sda")

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})

	When(" 2 devices are included ", func() {
		err = ndmutil.InitMinikube()
		Expect(err).NotTo(HaveOccurred())

		configMap = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
		configMap.SetIncludePath("/dev/sda", "/dev/vda")
		ndmutil.ReplaceAndApplyConfig(configMap)
		// It waits till node-disk-manager is ready or timeout reached
		err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		It("have both `/dev/sda` and `/dev/vda` in the pod", func() {
			matched, err := ndmutil.MatchNDMDeviceList(false, "/dev/sda", "/dev/vda")

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})

	When(" 2 devices are excluded ", func() {
		err = ndmutil.InitMinikube()
		Expect(err).NotTo(HaveOccurred())

		configMap = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
		configMap.SetExcludePath("/dev/sda", "/dev/vda")
		ndmutil.ReplaceAndApplyConfig(configMap)
		// It waits till node-disk-manager is ready or timeout reached
		err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		It("doesn't have either `/dev/sda` or `/dev/vda` in the pod", func() {
			matched, err := ndmutil.MatchNDMDeviceList(true, "/dev/sda", "/dev/vda")

			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(BeTrue())
		})
	})

})
