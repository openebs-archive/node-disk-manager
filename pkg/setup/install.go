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

import "fmt"

// Install installs the components based on configuration provided
func (sc Config) Install() error {

	var err error
	// create CRDs
	if err = sc.createBlockDeviceCRD(); err != nil {
		return fmt.Errorf("block device CRD creation failed : %v", err)
	}
	if err = sc.createBlockDeviceClaimCRD(); err != nil {
		return fmt.Errorf("block device claim CRD creation failed : %v", err)
	}

	return nil
}
