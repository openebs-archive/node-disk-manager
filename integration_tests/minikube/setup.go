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
	"strings"
)

// Status returns the status of the minikube cluster.
// The map contains status of each component.
// apiserver, kubelet etc.
func (minikube Minikube) Status() (map[string]string, error) {
	command := minikube.Command + " status"
	output, err := execCommand(command)
	if err != nil {
		return nil, fmt.Errorf("error getting status of minikube. Error: %v", err)
	}
	status := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		keyval := strings.SplitN(line, ":", 2)
		if len(keyval) == 1 {
			status[strings.TrimSpace(keyval[0])] = ""
		} else {
			status[strings.TrimSpace(keyval[0])] = strings.TrimSpace(keyval[1])
		}
	}
	return status, nil
}
