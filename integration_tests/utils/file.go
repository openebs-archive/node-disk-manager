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

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GetYAMLString gets the yaml-string from the given YAML file
func GetYAMLString(fileName string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// GetHomeDir gets the home directory for the system.
// It is required to locate the .kube/config file
func GetHomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}

	return "", fmt.Errorf("Not able to locate home directory")
}

// GetConfigPath returns the filepath of kubeconfig file
func GetConfigPath() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	kubeConfigPath := home + "/.kube/config"
	return kubeConfigPath, nil
}
