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

package kubernetes

import (
	api "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
)

func createFakeBlockDevice(uuid string) *blockdevice.BlockDevice {

	bd := &blockdevice.BlockDevice{
		Identifier: blockdevice.Identifier{
			UUID: uuid,
		},
	}
	// init all fields which require memory
	bd.NodeAttributes = make(map[string]string)
	return bd
}

func createFakeBlockDeviceAPI(name string) *api.BlockDevice {
	bdAPI := &api.BlockDevice{}
	bdAPI.Name = name
	bdAPI.Labels = make(map[string]string)
	return bdAPI
}
