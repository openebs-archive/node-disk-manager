package commonresource

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/CITF"
)

// CitfInstance is instance of CITF which will be used throughout the test
var CitfInstance citf.CITF

func init() {
	RegisterFailHandler(Fail)
	var err error
	CitfInstance, err = citf.NewCITF("")
	Expect(err).NotTo(HaveOccurred())
}
