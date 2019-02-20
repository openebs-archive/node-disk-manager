package commonresource

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	citf "github.com/openebs/CITF"
	citfoptions "github.com/openebs/CITF/citf_options"
)

// CitfInstance is instance of CITF which will be used throughout the test
var CitfInstance citf.CITF

func init() {
	RegisterFailHandler(Fail)
	var err error
	CitfInstance, err = citf.NewCITF(citfoptions.CreateOptionsIncludeAll(""))
	Expect(err).NotTo(HaveOccurred())
}
