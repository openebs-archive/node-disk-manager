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
	"context"
	"io"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/node-disk-manager/integration_tests/k8s"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sGoClient *kubernetes.Clientset

func TestNDM(t *testing.T) {
	// Setup go client for fetching pod logs on failure
	kubeConfigPath, err := utils.GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	goClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	k8sGoClient = goClient

	// Custom fail handler to print pod logs on failure
	RegisterFailHandler(NdmFail)
	RunSpecs(t, "Integration Test Suite")
}

// Initialize the suite
var _ = BeforeSuite(func() {
	// Create the client set
	c, err := k8s.GetClientSet()
	Expect(err).NotTo(HaveOccurred())

	// Create the crds
	err = c.CreateNDMCRDs()
	Expect(err).NotTo(HaveOccurred())

	err = c.CreateOpenEBSNamespace()
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

	err = c.DeleteOpenEBSNamespace()
	Expect(err).NotTo(HaveOccurred())
})

func NdmFail(message string, callerSkip ...int) {
	printNDMPodLogs()
	Fail(message, callerSkip...)
}

func printNDMPodLogs() {
	podInterface := k8sGoClient.CoreV1().Pods(k8s.Namespace)
	podList, err := podInterface.List(context.TODO(),
		metav1.ListOptions{LabelSelector: "name=" + DaemonSetPodPrefix})
	if err != nil {
		GinkgoWriter.Write([]byte("could not fetch pod logs: " + err.Error()))
		return
	}

	if len(podList.Items) == 0 {
		GinkgoWriter.Write([]byte("could not find ndm pod"))
		return
	}

	if len(podList.Items) > 1 {
		GinkgoWriter.Write(
			[]byte("found more than one ndm pods. will pick the first pod in the list"))
	}

	podName := podList.Items[0].Name
	podLogs, err := podInterface.GetLogs(podName,
		&corev1.PodLogOptions{}).Stream(context.TODO())
	if err != nil {
		GinkgoWriter.Write([]byte("could not fetch pod logs: " + err.Error()))
		return
	}
	defer podLogs.Close()
	GinkgoWriter.Write([]byte("------------------ NDM Logs ------------------\n"))
	_, err = io.Copy(GinkgoWriter, podLogs)
	if err != nil {
		GinkgoWriter.Write([]byte("could not write pod logs: " + err.Error()))
		return
	}
	GinkgoWriter.Write([]byte("-------------------- END --------------------\n"))
}
