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
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"time"
)

const (
	// CommandName is the command to execute the minikube binary
	CommandName = "minikube"
	// Running is the active/running status of minikube cluster
	Running = "Running"
)

var runCommand = utils.RunCommandWithSudo
var execCommand = utils.ExecCommandWithSudo

// Minikube contains the objects to be used while interfacing
// with a minikube cluster
type Minikube struct {
	// the command to execute for minikube cluster
	Command string

	// timeout duration for all minikube operations
	Timeout time.Duration

	// wait time for minikube operation. If the command does not return
	// we wait continously till timeout is reached
	WaitTime time.Duration
}

// NewMinikube returns a new minikube struct with the command to execute
// and the default wait-timeout
func NewMinikube() Minikube {
	return Minikube{
		Command:  CommandName,
		Timeout:  time.Minute,
		WaitTime: time.Second,
	}
}
