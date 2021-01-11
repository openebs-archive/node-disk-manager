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
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestIsInstallCRDEnabled(t *testing.T) {
	tests := map[string]struct {
		setEnv   bool
		envValue string
		want     bool
	}{
		"when INSTALL_CRD_ENV is set to true": {
			setEnv:   true,
			envValue: "true",
			want:     true,
		},
		"when INSTALL_CRD_ENV is set to false": {
			setEnv:   true,
			envValue: "false",
		},
		"when INSTALL_CRD_ENV is not set": {
			setEnv: false,
			want:   true,
		},
		"when INSTALL_CRD is set to empty": {
			setEnv:   true,
			envValue: "",
			want:     true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(INSTALL_CRD_ENV, tt.envValue)
			}
			assert.Equal(t, tt.want, IsInstallCRDEnabled())
			_ = os.Unsetenv(INSTALL_CRD_ENV)
		})
	}
}

func TestGetOpenEBSImagePullSecrets(t *testing.T) {
	tests := map[string]struct {
		envValue string
		secret   []v1.LocalObjectReference
	}{
		"empty variable": {
			envValue: "",
			secret:   []v1.LocalObjectReference{},
		},
		"single value": {
			envValue: "image-pull-secret",
			secret:   []v1.LocalObjectReference{{Name: "image-pull-secret"}},
		},
		"multiple value": {
			envValue: "image-pull-secret,secret-1",
			secret:   []v1.LocalObjectReference{{Name: "image-pull-secret"}, {Name: "secret-1"}},
		},
		"whitespaces": {
			envValue: " ",
			secret:   []v1.LocalObjectReference{},
		},
		"single value with whitespaces": {
			envValue: " docker-secret ",
			secret:   []v1.LocalObjectReference{{Name: "docker-secret"}},
		},
		"multiple value with whitespaces": {
			envValue: " docker-secret, image-pull-secret ",
			secret:   []v1.LocalObjectReference{{Name: "docker-secret"}, {Name: "image-pull-secret"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(IMAGE_PULL_SECRETS_ENV, tt.envValue)
			got := GetOpenEBSImagePullSecrets()
			assert.Equal(t, tt.secret, got)
			os.Unsetenv(IMAGE_PULL_SECRETS_ENV)
		})
	}
}
