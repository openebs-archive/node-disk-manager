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
