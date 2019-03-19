package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Get the yaml-string from the given YAML file
func GetYAMLString(fileName string) (string, error) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// Get the home directory for the system.
// It is required to locate the .kube/config file
func GetHomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}

	return "", fmt.Errorf("Not able to locate home directory")
}

// Get the filepath of kubeconfig file
func GetConfigPath() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	kubeConfigPath := home + "/.kube/config"
	return kubeConfigPath, nil
}
