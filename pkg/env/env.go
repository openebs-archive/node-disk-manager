/*
Copyright 2020 The OpenEBS Authors

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

package env

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	// IMAGE_PULL_SECRETS_ENV is the environment variable used to pass the image pull secrets
	IMAGE_PULL_SECRETS_ENV = "OPENEBS_IO_IMAGE_PULL_SECRETS"

	// WATCH_NAMESPACE is the namespace to watch for resources
	WATCH_NAMESPACE_ENV = "WATCH_NAMESPACE"
)

// GetOpenEBSImagePullSecrets is used to get the image pull secrets from the environment variable
func GetOpenEBSImagePullSecrets() []v1.LocalObjectReference {
	secrets := strings.TrimSpace(os.Getenv(IMAGE_PULL_SECRETS_ENV))

	list := make([]v1.LocalObjectReference, 0)

	if len(secrets) == 0 {
		return list
	}
	arr := strings.Split(secrets, ",")
	for _, item := range arr {
		if len(item) > 0 {
			l := v1.LocalObjectReference{Name: strings.TrimSpace(item)}
			list = append(list, l)
		}
	}
	return list
}

// GetWatchNamespace gets the namespace to watch for resources
func GetWatchNamespace() (string, error) {
	ns := strings.TrimSpace(os.Getenv(WATCH_NAMESPACE_ENV))
	if len(ns) == 0 {
		return "", fmt.Errorf("WATCH_NAMESPACE env not set")
	}
	return ns, nil
}
