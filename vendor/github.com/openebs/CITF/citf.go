/*
Copyright 2018 The OpenEBS Authors.
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

package citf

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/CITF/common"
	"github.com/openebs/CITF/config"
	"github.com/openebs/CITF/environments"
	"github.com/openebs/CITF/environments/docker"
	"github.com/openebs/CITF/environments/minikube"
	"github.com/openebs/CITF/utils/k8s"
	"github.com/openebs/CITF/utils/log"
)

// CITF is a struct which will be the driver for all functionalities of this framework
type CITF struct {
	Environment  environments.Environment
	K8S          k8s.K8S
	Docker       docker.Docker
	DebugEnabled bool
	Logger       log.Logger
}

// NewCITF returns CITF struct. One need this in order to use any functionality of this framework.
func NewCITF(confFilePath string) (CITF, error) {
	var environment environments.Environment
	if err := config.LoadConf(confFilePath); err != nil {
		// Log this here
		// Here, we don't want to return fatal error since we want to continue
		// executing the function with default configuration even if it fails
		glog.Errorf("error loading config file. Error: %+v", err)
	}

	switch config.Environment() {
	case common.Minikube:
		environment = minikube.NewMinikube()
	default:
		// Exit with Error
		return CITF{}, fmt.Errorf("platform: %q is not suppported by CITF", config.Environment())
	}

	k8sInstance, err := k8s.NewK8S()
	if err != nil {
		return CITF{}, err
	}

	return CITF{
		K8S:          k8sInstance,
		Environment:  environment,
		Docker:       docker.NewDocker(),
		DebugEnabled: config.Debug(),
		Logger:       log.Logger{},
	}, nil
}
