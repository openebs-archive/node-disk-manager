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

package filter

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeviceValidityFilterExclude(t *testing.T) {
	tests := map[string]struct {
		blockDevice *blockdevice.BlockDevice
		want        bool
	}{
		"valid BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 1024,
				},
			},
			want: true,
		},
		"invalid Path in BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 1024,
				},
			},
			want: false,
		},
		"invalid Capacity in BlockDevice": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
		"invalid Capacity and DevPath": {
			blockDevice: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dvf := deviceValidityFilter{}
			dvf.Start()
			assert.Equal(t, test.want, dvf.Exclude(test.blockDevice))
		})
	}
}

func TestIsValidDevPath(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid dev path": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			want: true,
		},
		"invalid dev path": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "",
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isValidDevPath(test.bd))
		})
	}
}

func Test_isValidCapacity(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid capacity": {
			bd: &blockdevice.BlockDevice{
				Capacity: blockdevice.CapacityInformation{
					Storage: 102400,
				},
			},
			want: true,
		},
		"invalid capacity": {
			bd: &blockdevice.BlockDevice{
				Capacity: blockdevice.CapacityInformation{
					Storage: 0,
				},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isValidCapacity(test.bd))
		})
	}
}

func TestIsValidPartition(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid blockdevice of type partition": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionEntryUUID: "065e2357-05",
				},
			},
			want: true,
		},
		"invalid blockdevice of type partition": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isValidPartition(tt.bd)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidDMDevice(t *testing.T) {
	tests := map[string]struct {
		bd   *blockdevice.BlockDevice
		want bool
	}{
		"valid blockdevice of type device mapper": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/dm-0",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeLVM,
				},
				DMInfo: blockdevice.DeviceMapperInformation{
					DMUUID: "LVM-j2xmqvbcVWBQK9Jdttte3CyeVTGgxtVV5VcCi3nxdwihZDxSquMOBaGL5eymBNvk",
				},
			},
			want: true,
		},
		"invalid blockdevice of type device mapper": {
			bd: &blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/dm-1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeLVM,
				},
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isValidDMDevice(tt.bd)
			assert.Equal(t, tt.want, got)
		})
	}
}
