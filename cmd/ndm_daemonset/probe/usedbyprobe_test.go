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

package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBlockDeviceZFSPartition(t *testing.T) {
	tests := map[string]struct {
		bd    blockdevice.BlockDevice
		want  string
		want1 bool
	}{
		"blockdevice has 2 partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/sda1", "/dev/sda2"},
				},
			},
			want:  "",
			want1: false,
		},
		"blockdevice has 2 zfs partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/sda1", "/dev/sda9"},
				},
			},
			want:  "/dev/sda1",
			want1: true,
		},
		"nvme blockdevice has 2 partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/nvme0n1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/nvme0n1p1", "/dev/nvme0n1p2"},
				},
			},
			want:  "",
			want1: false,
		},
		"nvme blockdevice has 2 zfs partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/nvme0n1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/nvme0n1p1", "/dev/nvme0n1p9"},
				},
			},
			want:  "/dev/nvme0n1p1",
			want1: true,
		},
		"blockdevice has multiple partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/sda1", "/dev/sda2", "/dev/sda9"},
				},
			},
			want:  "",
			want1: false,
		},
		"nvme blockdevice has multiple partitions": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/nvme0n1",
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Partitions: []string{"/dev/nvme0n1p1", "/dev/nvme0n1p2", "/dev/nvme0n1p9"},
				},
			},
			want:  "",
			want1: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := getBlockDeviceZFSPartition(tt.bd)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
