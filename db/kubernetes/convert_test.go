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
	"testing"

	api "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
)

func Test_convert_BlockDeviceAPI_To_BlockDevice(t *testing.T) {
	type args struct {
		in      *api.BlockDevice
		wantOut *blockdevice.BlockDevice
	}

	fakeBDName := "my-fake-bd"
	fakeHostName := "fake-hostname"
	fakeNodeName := "fake-machine"
	fakeDevicePath := "/dev/sdf1"
	fileSystem := "ext4"
	mountPoint := "/mnt/media"
	deviceType := blockdevice.SparseBlockDeviceType

	// building the blockdevice API object
	in1 := createFakeBlockDeviceAPI(fakeBDName)
	in1.Labels[KubernetesHostNameLabel] = fakeHostName
	in1.Spec.NodeAttributes.NodeName = fakeNodeName
	in1.Spec.Path = fakeDevicePath
	in1.Spec.FileSystem.Type = fileSystem
	in1.Spec.FileSystem.Mountpoint = mountPoint
	in1.Spec.Details.DeviceType = deviceType
	in1.Status.State = api.BlockDeviceState(blockdevice.Active)
	in1.Status.ClaimState = api.DeviceClaimState(blockdevice.Claimed)

	// building the core blockdevice object
	out1 := createFakeBlockDevice(fakeBDName)
	out1.NodeAttributes[blockdevice.HostName] = fakeHostName
	out1.NodeAttributes[blockdevice.NodeName] = fakeNodeName
	out1.DevPath = fakeDevicePath
	out1.FSInfo.FileSystem = fileSystem
	out1.FSInfo.MountPoint = append(out1.FSInfo.MountPoint, mountPoint)
	out1.DeviceAttributes.DeviceType = blockdevice.SparseBlockDeviceType
	out1.Status.State = blockdevice.Active
	out1.Status.ClaimPhase = blockdevice.Claimed

	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"converting block device k8s resource to BlockDevice": {
			args: args{
				in:      in1,
				wantOut: out1,
			},
			wantErr: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotOut := &blockdevice.BlockDevice{}
			err := convertBlockDeviceAPIToBlockDevice(test.args.in, gotOut)
			assert.Equal(t, test.args.wantOut, gotOut)
			assert.Equal(t, test.wantErr, err != nil)
		})
	}
}
