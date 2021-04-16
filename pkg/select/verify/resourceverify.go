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

package verify

import (
	"fmt"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// GetRequestedCapacity gets the requested capacity from the BlockDeviceClaim
// It returns an error if the Quantity cannot be parsed
func GetRequestedCapacity(list v1.ResourceList) (int64, error) {

	resourceCapacity := list[apis.ResourceStorage]
	// Check if deviceClaim has valid capacity request
	capacity, err := (&resourceCapacity).AsInt64()
	if !err || capacity <= 0 {
		return 0, fmt.Errorf("invalid capacity requested, %v", err)
	}
	return capacity, nil
}
