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

	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	// INSTALL_CRD_ENV is the environment variable used to check if
	// CRDs need to be installed by NDM or not.
	INSTALL_CRD_ENV = "OPENEBS_IO_INSTALL_CRD"

	// installCRDEnvDefaultValue is the default value for the INSTALL_CRD_ENV
	installCRDEnvDefaultValue = true
)

// IsInstallCRDEnabled is used to check whether the CRDs need to be installed
func IsInstallCRDEnabled() bool {
	val := os.Getenv(INSTALL_CRD_ENV)

	// if empty return the default value
	if len(val) == 0 {
		return installCRDEnvDefaultValue
	}

	return util.CheckTruthy(val)
}
