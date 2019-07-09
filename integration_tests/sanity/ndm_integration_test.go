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
	. "github.com/openebs/node-disk-manager/integration_tests/minikube"
	"testing"
)

var (
	minikube = NewMinikube()
)

func TestNDM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

// Initialize the minikube cluster
var _ = BeforeSuite(func() {
	var err error
	Expect(minikube.IsUpAndRunning()).To(BeTrue())
	err = minikube.WaitForMinikubeToBeReady()
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Verify Kubernetes Cluster Setup", func() {
	var err error
	Context("Initially, we check minikube status", func() {
		It("should be running", func() {
			Expect(minikube.IsUpAndRunning()).To(BeTrue())
		})
	})
	Context("We check for generated Kube Config", func() {
		_, err = k8s.GetClientSet()
		It("should be able to generate ClientSet", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
