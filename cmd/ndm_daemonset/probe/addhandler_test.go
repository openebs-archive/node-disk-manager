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
	"reflect"
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestAddBlockDeviceToHierarchyCache(t *testing.T) {
	tests := map[string]struct {
		cache     blockdevice.Hierarchy
		bd        blockdevice.BlockDevice
		wantCache blockdevice.Hierarchy
		wantOk    bool
	}{
		"empty cache": {
			cache: make(blockdevice.Hierarchy),
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			wantOk: false,
		},
		"cache with same device already existing": {
			cache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
					SysPath: "/sys/class/block/sda",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						SysPath: "/sys/class/block/sda",
						DevPath: "/dev/sda",
					},
				},
			},
			wantOk: true,
		},
		"cache with different device existing": {
			cache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
			},
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb",
				},
			},
			wantCache: map[string]blockdevice.BlockDevice{
				"/dev/sda": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sda",
					},
				},
				"/dev/sdb": {
					Identifier: blockdevice.Identifier{
						DevPath: "/dev/sdb",
					},
				},
			},
			wantOk: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: &controller.Controller{
					BDHierarchy: tt.cache,
				},
			}
			gotOk := pe.addBlockDeviceToHierarchyCache(tt.bd)
			assert.Equal(t, tt.wantCache, pe.Controller.BDHierarchy)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestDeviceInUseByMayastor(t *testing.T) {
	type fields struct {
		Controller *controller.Controller
	}
	type args struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: tt.fields.Controller,
			}
			got, err := pe.deviceInUseByMayastor(tt.args.bd, tt.args.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("deviceInUseByMayastor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("deviceInUseByMayastor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceInUseByZFSLocalPV(t *testing.T) {
	type fields struct {
		Controller *controller.Controller
	}
	type args struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: tt.fields.Controller,
			}
			got, err := pe.deviceInUseByZFSLocalPV(tt.args.bd, tt.args.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("deviceInUseByZFSLocalPV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("deviceInUseByZFSLocalPV() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleUnmanagedDevices(t *testing.T) {
	type fields struct {
		Controller *controller.Controller
	}
	type args struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: tt.fields.Controller,
			}
			got, err := pe.handleUnmanagedDevices(tt.args.bd, tt.args.bdAPIList)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUnmanagedDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("handleUnmanagedDevices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsParentDeviceInUse(t *testing.T) {
	cache := map[string]blockdevice.BlockDevice{
		"/dev/sda": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent:     "",
				Partitions: []string{"/dev/sda1", "/dev/sda2"},
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypeDisk,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: false,
			},
		},
		"/dev/sda1": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda1",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "/dev/sda",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypePartition,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: true,
			},
		},
		"/dev/sda2": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda2",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "/dev/sda",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypePartition,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: false,
			},
		},
		"/dev/sdb": {
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sdb",
			},
			DependentDevices: blockdevice.DependentBlockDevices{
				Parent: "",
			},
			DeviceAttributes: blockdevice.DeviceAttribute{
				DeviceType: blockdevice.BlockDeviceTypeDisk,
			},
			DevUse: blockdevice.DeviceUsage{
				InUse: true,
			},
		},
	}
	pe := &ProbeEvent{
		Controller: &controller.Controller{
			BDHierarchy: cache,
		},
	}
	tests := map[string]struct {
		bd      blockdevice.BlockDevice
		want    bool
		wantErr bool
	}{
		"check for existing parent device": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypeDisk,
				},
			},
			want:    false,
			wantErr: false,
		},
		"check for partition that is in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sda1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sda",
				},
			},
			want:    false,
			wantErr: false,
		},
		"check for parent device in use": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdb1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sdb",
				},
			},
			want:    true,
			wantErr: false,
		},
		"non existent parent device": {
			bd: blockdevice.BlockDevice{
				Identifier: blockdevice.Identifier{
					DevPath: "/dev/sdc1",
				},
				DeviceAttributes: blockdevice.DeviceAttribute{
					DeviceType: blockdevice.BlockDeviceTypePartition,
				},
				DependentDevices: blockdevice.DependentBlockDevices{
					Parent: "/dev/sdc",
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := pe.isParentDeviceInUse(tt.bd)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, gotErr != nil)
		})
	}
}

func TestIsBDWithFsUuidExists(t *testing.T) {
	type args struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
	}
	tests := []struct {
		name string
		args args
		want *apis.BlockDevice
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExistingBDWithFsUuid(tt.args.bd, tt.args.bdAPIList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getExistingBDWithFsUuid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBDWithPartitionUUIDExists(t *testing.T) {
	type args struct {
		bd        blockdevice.BlockDevice
		bdAPIList *apis.BlockDeviceList
	}
	tests := []struct {
		name string
		args args
		want *apis.BlockDevice
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExistingBDWithPartitionUUID(tt.args.bd, tt.args.bdAPIList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getExistingBDWithPartitionUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
