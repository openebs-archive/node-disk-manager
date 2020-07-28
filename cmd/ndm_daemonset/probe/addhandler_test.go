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

func TestProbeEvent_deviceInUseByMayastor(t *testing.T) {
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

func TestProbeEvent_deviceInUseByZFSLocalPV(t *testing.T) {
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

func TestProbeEvent_handleUnmanagedDevices(t *testing.T) {
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
	tests := map[string]struct {
		bd      blockdevice.BlockDevice
		cache   blockdevice.Hierarchy
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pe := &ProbeEvent{
				Controller: &controller.Controller{
					BDHierarchy: tt.cache,
				},
			}
			got, gotErr := pe.isParentDeviceInUse(tt.bd)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, gotErr)
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
