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
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	fakeWWN := "50E5495131BBB060892FBC8E"
	fakeSerial := "CT500MX500SSD1"
	fakeFileSystemUUID := "149108ca-f404-4556-a263-04943e6cb0b3"
	fakePartitionUUID := "065e2357-05"
	tests := map[string]struct {
		bd       blockdevice.BlockDevice
		wantUUID string
		wantOk   bool
	}{
		"deviceType-disk with WWN": {
			bd: blockdevice.BlockDevice{
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					WWN:        fakeWWN,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeWWN),
			wantOk:   true,
		},
		"deviceType-disk with WWN and serial": {
			bd: blockdevice.BlockDevice{
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					WWN:        fakeWWN,
					Serial:     fakeSerial,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeWWN+fakeSerial),
			wantOk:   true,
		},
		"deviceType-disk with a filesystem and no wwn": {
			bd: blockdevice.BlockDevice{
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakeFileSystemUUID,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeFileSystemUUID),
			wantOk:   true,
		},
		"deviceType-disk with a filesystem and wwn": {
			bd: blockdevice.BlockDevice{
				FSInfo: blockdevice.FileSystemInformation{
					FileSystemUUID: fakeFileSystemUUID,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
					WWN:        fakeWWN,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeWWN),
			wantOk:   true,
		},
		"deviceType-partition with wwn on the disk": {
			bd: blockdevice.BlockDevice{
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
					WWN:        fakeWWN,
				},
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionEntryUUID: fakePartitionUUID,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakePartitionUUID),
			wantOk:   true,
		},
		"deviceType-disk with no wwn or filesystem": {
			bd: blockdevice.BlockDevice{
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			wantUUID: "",
			wantOk:   false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotUUID, gotOk := generateUUID(tt.bd)
			assert.Equal(t, tt.wantUUID, gotUUID)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
