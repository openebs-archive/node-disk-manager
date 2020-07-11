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
