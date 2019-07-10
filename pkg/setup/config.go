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

package setup

import (
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"time"
)

const (
	// CRDRetryInterval is the duration to wait while retrying for CRDs
	CRDRetryInterval = 4 * time.Second
	// CRDTimeout is the duration after which retry timeouts
	CRDTimeout = 30 * time.Second
)

// Config defines the config for installation
type Config struct {
	apiExtClient *apiextclient.Clientset
}

// NewInstallSetup creates the installation setup struct which
// can be used for generating the config and client used during installation
func NewInstallSetup(config *rest.Config) (*Config, error) {
	setupConfig := &Config{}
	client, err := apiextclient.NewForConfig(config)
	if err != nil {
		return setupConfig, nil
	}
	setupConfig.apiExtClient = client
	return setupConfig, nil
}
