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
	"os"
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/features"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	fakeWWN := "50E5495131BBB060892FBC8E"
	fakeSerial := "CT500MX500SSD1"
	fakeFileSystemUUID := "149108ca-f404-4556-a263-04943e6cb0b3"
	fakePartitionUUID := "065e2357-05"
	fakePartitionTableUUID := "6f479331-dad4-4ccb-b146-5c359c55399b"
	fakeLVM_DM_UUID := "LVM-j2xmqvbcVWBQK9Jdttte3CyeVTGgxtVV5VcCi3nxdwihZDxSquMOBaGL5eymBNvk"
	fakeCRYPT_DM_UUID := "CRYPT-LUKS1-f4608c76343d4b5badaf6651d32f752b-backup"
	loopDevicePath := "/dev/loop98"
	hostName, _ := os.Hostname()
	features.FeatureGates.SetFeatureFlag([]string{
		"GPTBasedUUID=1",
		"PartitionTableUUID=1",
	})
	tests := map[string]struct {
		bd       blockdevice.BlockDevice
		wantUUID string
		wantOk   bool
	}{
		"debiceType-disk with PartitionTableUUID": {
			bd: blockdevice.BlockDevice{
				PartitionInfo: blockdevice.PartitionInformation{
					PartitionTableType: "gpt",
					PartitionTableUUID: fakePartitionTableUUID,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakePartitionTableUUID),
			wantOk:   true,
		},
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
		"deviceType-lvm device": {
			bd: blockdevice.BlockDevice{
				DMInfo: blockdevice.DeviceMapperInformation{
					DMUUID: fakeLVM_DM_UUID,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeLVM,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeLVM_DM_UUID),
			wantOk:   true,
		},
		"deviceType-crypt device": {
			bd: blockdevice.BlockDevice{
				DMInfo: blockdevice.DeviceMapperInformation{
					DMUUID: fakeCRYPT_DM_UUID,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeCrypt,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeCRYPT_DM_UUID),
			wantOk:   true,
		},
		"deviceType-loop device": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: loopDevicePath,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeLoop,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(hostName+loopDevicePath),
			wantOk:   true,
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

func TestGenerateLegacyUUID(t *testing.T) {
	fakePath := "/dev/sda"
	fakeWWN := "50E5495131BBB060892FBC8E"
	fakeSerial := "CT500MX500SSD1"
	fakeModel := "DataTraveler_3.0"
	fakeVendor := "Kingston"
	hostname, _ := os.Hostname()
	tests := map[string]struct {
		bd       blockdevice.BlockDevice
		wantUUID string
		wantOk   bool
	}{
		"NonLocal Disk Model with wwn/vendor/model/serial": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: fakePath,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					IDType: "disk",
					WWN:    fakeWWN,
					Vendor: fakeVendor,
					Model:  fakeModel,
					Serial: fakeSerial,
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeWWN+fakeModel+fakeSerial+fakeVendor),
			wantOk:   false,
		},
		"local disk model with wwn": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: fakePath,
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					WWN:   fakeWWN,
					Model: "Virtual_disk",
				},
			},
			wantUUID: blockdevice.BlockDevicePrefix + util.Hash(fakeWWN+"Virtual_disk"+hostname+fakePath),
			wantOk:   true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotUUID, gotOk := generateLegacyUUID(tt.bd)
			assert.Equal(t, tt.wantUUID, gotUUID)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
