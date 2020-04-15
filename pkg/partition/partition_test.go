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

package partition

import (
	"github.com/diskfs/go-diskfs/partition/gpt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePartitionTable(t *testing.T) {
	tests := map[string]struct {
		actualDisk             Disk
		expectedPartitionTable *gpt.Table
		wantErr                bool
	}{
		"disk size is zero": {
			actualDisk: Disk{
				DevPath:          "/dev/sda",
				DiskSize:         0,
				LogicalBlockSize: 0,
				table:            nil,
			},
			expectedPartitionTable: nil,
			wantErr:                true,
		},
		"disk with zero block size": {
			actualDisk: Disk{
				DevPath:          "/dev/sda",
				DiskSize:         500107862016,
				LogicalBlockSize: 0,
				table:            nil,
			},
			expectedPartitionTable: &gpt.Table{
				LogicalSectorSize: 512,
				ProtectiveMBR:     true,
			},
			wantErr: false,
		},
		"disk with 4k block size": {
			actualDisk: Disk{
				DevPath:          "/dev/sda",
				DiskSize:         500107862016,
				LogicalBlockSize: 4096,
				table:            nil,
			},
			expectedPartitionTable: &gpt.Table{
				LogicalSectorSize: 4096,
				ProtectiveMBR:     true,
			},
			wantErr: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.actualDisk.createPartitionTable(); (err != nil) != test.wantErr {
				t.Errorf("CreatePartitionTable() error = %v, wantErr %v", err, test.wantErr)
			}
			assert.Equal(t, test.actualDisk.table, test.expectedPartitionTable)
		})
	}
}

func TestAddPartition(t *testing.T) {
	tests := map[string]struct {
		actualDisk             Disk
		expectedPartitionTable *gpt.Table
		wantErr                bool
	}{
		"465GiB HDD with 512 block size": {
			actualDisk: Disk{
				DevPath:          "/dev/sda",
				DiskSize:         500107862016,
				LogicalBlockSize: 512,
				table:            &gpt.Table{},
			},
			expectedPartitionTable: &gpt.Table{
				Partitions: []*gpt.Partition{
					{
						Start: 2048,
						End:   976773134,
						Type:  gpt.LinuxFilesystem,
					},
				},
			},
			wantErr: false,
		},
		"375 GiB SSD with 4k block size": {
			actualDisk: Disk{
				DevPath:          "/dev/sda",
				DiskSize:         402653184000,
				LogicalBlockSize: 4096,
				table:            &gpt.Table{},
			},
			expectedPartitionTable: &gpt.Table{
				Partitions: []*gpt.Partition{
					{
						Start: 256,
						End:   98303994,
						Type:  gpt.LinuxFilesystem,
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.actualDisk.addPartition(); (err != nil) != test.wantErr {
				t.Errorf("AddPartition() error = %v, wantErr %v", err, test.wantErr)
			}
			assert.Equal(t, test.actualDisk.table, test.expectedPartitionTable)
		})
	}
}
