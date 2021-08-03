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

func convertBlockDeviceAPIListToBlockDeviceList(in *api.BlockDeviceList, out *[]blockdevice.BlockDevice) error {
	var err error
	var bd blockdevice.BlockDevice
	for i := range in.Items {
		err = convertBlockDeviceAPIToBlockDevice(&in.Items[i], &bd)
		if err != nil {
			return err
		}
		*out = append(*out, bd)
	}
	return nil
}

func convertBlockDeviceAPIToBlockDevice(in *api.BlockDevice, out *blockdevice.BlockDevice) error {
	out.UUID = in.Name

	//labels
	out.NodeAttributes = make(blockdevice.NodeAttribute)
	out.NodeAttributes[blockdevice.HostName] = in.Labels[KubernetesHostNameLabel]
	out.NodeAttributes[blockdevice.NodeName] = in.Spec.NodeAttributes.NodeName

	//spec
	out.DevPath = in.Spec.Path
	out.FSInfo.FileSystem = in.Spec.FileSystem.Type

	// currently only the first mount point is filled in. When API is changed, multiple mount points
	// will be added.
	out.FSInfo.MountPoint = append(out.FSInfo.MountPoint, in.Spec.FileSystem.Mountpoint)
	out.DeviceAttributes.DeviceType = in.Spec.Details.DeviceType

	//status
	out.Status.State = string(in.Status.State)
	out.Status.ClaimPhase = string(in.Status.ClaimState)

	return nil
}
