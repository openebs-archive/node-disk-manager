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

package minikube

import (
	"fmt"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"os"
	"time"
)

// IsUpAndRunning checks if all components in the minikube cluster is
// in running state. If the api-server is Running we can proceed
// with further operations on minikube
func (minikube Minikube) IsUpAndRunning() bool {
	startTime := time.Now()
	var status map[string]string
	var err error
	for time.Since(startTime) < minikube.Timeout {
		status, err = minikube.Status()
		if err != nil {
			time.Sleep(minikube.WaitTime)
		}
		// loop through until apiserver is in Running state
		if state, ok := status["apiserver"]; !ok || state != Running {
			time.Sleep(minikube.WaitTime)
		} else if state == Running {
			return true
		}
	}
	return false
}

// WaitForMinikubeToBeReady waits for a fixed time or till the kube-config
// file is available
func (minikube Minikube) WaitForMinikubeToBeReady() error {
	startTime := time.Now()
	configPath, _ := utils.GetConfigPath()
	for time.Since(startTime) < minikube.Timeout {
		_, err := os.Stat(configPath)
		if os.IsNotExist(err) {
			fmt.Println("waiting for kubeconfig file to be generated.")
			time.Sleep(minikube.WaitTime)
		} else {
			return nil
		}
	}
	if time.Since(startTime) >= minikube.Timeout {
		return fmt.Errorf("Kubeconfig file not generated")
	}
	return nil
}
