package integrationtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_test/ndm_util"
	"time"
)

var _ = Describe("Path filter integration Test", func() {

	var configMapPatch ndmutil.ConfigMapPatch
	var err error

	Context("Path filter is disabled", func() {
		When("We disable the path filter", func() {
			err = ndmutil.InitEnvironment()
			Expect(err).NotTo(HaveOccurred())

			configMapPatch = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
			configMapPatch.SetPathFilter("false")

			ndmutil.ReplaceAndApplyConfig(configMapPatch)
			// It waits till node-disk-manager is ready or timeout reached
			err = ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
			Expect(err).NotTo(HaveOccurred())

			It("has matching disks inside node and the host", func() {

				matched, err := ndmutil.MatchDisksOutsideAndInside()

				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeTrue())
			})
		})
	})

	Context("When path filter is enabled", func() {

		When("We include a single device", func() {
			err = ndmutil.InitEnvironment()
			Expect(err).NotTo(HaveOccurred())

			configMapPatch = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
			configMapPatch.SetIncludePath("/dev/sda")
			ndmutil.ReplaceAndApplyConfig(configMapPatch)
			// It waits till node-disk-manager is ready or timeout reached
			err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
			Expect(err).NotTo(HaveOccurred())

			It("has only /dev/sda in the pod", func() {
				matched, err := ndmutil.MatchNDMDeviceList(false, "/dev/sda")

				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeTrue())
			})
		})

		When("We exclude a single device", func() {
			err = ndmutil.InitEnvironment()
			Expect(err).NotTo(HaveOccurred())

			configMapPatch = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
			configMapPatch.SetExcludePath("/dev/sda")
			ndmutil.ReplaceAndApplyConfig(configMapPatch)
			// It waits till node-disk-manager is ready or timeout reached
			err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
			Expect(err).NotTo(HaveOccurred())

			It("does not have /dev/sda in the pod", func() {
				matched, err := ndmutil.MatchNDMDeviceList(true, "/dev/sda")

				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeTrue())
			})
		})

		When("We include 2 devices", func() {
			err = ndmutil.InitEnvironment()
			Expect(err).NotTo(HaveOccurred())

			configMapPatch = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
			configMapPatch.SetIncludePath("/dev/sda", "/dev/vda")
			ndmutil.ReplaceAndApplyConfig(configMapPatch)
			// It waits till node-disk-manager is ready or timeout reached
			err := ndmutil.WaitTillNDMisUpOrTimeout(5 * time.Minute)
			Expect(err).NotTo(HaveOccurred())

			It("have both `/dev/sda` and `/dev/vda` in the pod", func() {
				matched, err := ndmutil.MatchNDMDeviceList(false, "/dev/sda", "/dev/vda")

				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeTrue())
			})
		})

		When("We exclude 2 devices", func() {
			err = ndmutil.InitEnvironment()
			Expect(err).NotTo(HaveOccurred())

			configMapPatch = ndmutil.GetNDMConfig(ndmutil.GetNDMOperatorFilePath())
			configMapPatch.SetExcludePath("/dev/sda", "/dev/vda")
			ndmutil.ReplaceAndApplyConfig(configMapPatch)
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
})
