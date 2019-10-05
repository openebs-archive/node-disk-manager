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
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"testing"
)

func TestNDM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

// Initialize the suite
var _ = BeforeSuite(func() {
	// Create the client set
	c, err := k8s.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	// Create service account and cluster roles required for NDM
	err = c.CreateNDMServiceAccount()
	Expect(err).NotTo(HaveOccurred())

	err = c.CreateNDMClusterRole()
	Expect(err).NotTo(HaveOccurred())

	err = c.CreateNDMClusterRoleBinding()
	Expect(err).NotTo(HaveOccurred())

	err = c.CreateNDMConfigMap()
	Expect(err).NotTo(HaveOccurred())

	err = c.CreateNDMOperatorDeployment()
	Expect(err).NotTo(HaveOccurred())

	// wait for all changes to happen
	k8s.WaitForStateChange()

})

// clean up all resources by NDM
var _ = AfterSuite(func() {
	c, err := k8s.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMClusterRoleBinding()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMClusterRole()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMServiceAccount()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMCRDs()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMConfigMap()
	Expect(err).NotTo(HaveOccurred())

	err = c.DeleteNDMOperatorDeployment()
	Expect(err).NotTo(HaveOccurred())
})
