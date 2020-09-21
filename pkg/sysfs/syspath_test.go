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

package sysfs

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

func TestGetParent(t *testing.T) {
	type fields struct {
		DeviceName string
		SysPath    string
	}
	tests := map[string]struct {
		fields           fields
		wantedDeviceName string
		wantOk           bool
	}{
		"[block] given block device is a parent": {
			fields: fields{
				DeviceName: "sda",
				SysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[block] given blockdevice is a partition": {
			fields: fields{
				DeviceName: "sda4",
				SysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4",
			},
			wantedDeviceName: "sda",
			wantOk:           true,
		},
		"[nvme] given blockdevice is a parent": {
			fields: fields{
				DeviceName: "nvme0n1",
				SysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[nvme] given blockdevice is a partition": {
			fields: fields{
				DeviceName: "nvme0n1p1",
				SysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1",
			},
			wantedDeviceName: "nvme0n1",
			wantOk:           true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := Device{}
			gotDeviceName, gotOk := s.getParent()
			assert.Equal(t, test.wantedDeviceName, gotDeviceName)
			assert.Equal(t, test.wantOk, gotOk)
		})
	}
}

func TestGetDeviceSysPath(t *testing.T) {
	sysFSDirectoryPath = "/tmp/sys/"

	pciPath := "devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/"

	// create top level sys directory
	os.MkdirAll(sysFSDirectoryPath, 0700)
	// create block directory
	os.MkdirAll(sysFSDirectoryPath+"class/block", 0700)
	// create devices directory
	os.MkdirAll(sysFSDirectoryPath+"devices", 0700)

	// create device directory
	os.MkdirAll(sysFSDirectoryPath+pciPath, 0700)
	os.Symlink(sysFSDirectoryPath+pciPath, sysFSDirectoryPath+"class/block/sda")

	tests := map[string]struct {
		devicePath string
		want       string
		wantErr    bool
	}{
		"devicenode name is used": {
			devicePath: "/dev/sda",
			want:       sysFSDirectoryPath + pciPath,
			wantErr:    false,
		},
		"actual syspath is used": {
			devicePath: sysFSDirectoryPath + pciPath,
			want:       sysFSDirectoryPath + pciPath,
			wantErr:    false,
		},
		"class/block path is used": {
			devicePath: sysFSDirectoryPath + "class/block/sda",
			want:       sysFSDirectoryPath + pciPath,
			wantErr:    false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotSysPath, err := getDeviceSysPath(tt.devicePath)
			if err != nil {
				t.Errorf("error getDeviceSysPath() for %s, error: %v", tt.devicePath, err)
			}
			assert.Equal(t, tt.want, gotSysPath)
		})
	}
}

func TestSysFsDeviceGetPartitions(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, got1 := s.getPartitions()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPartitions() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getPartitions() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSysFsDeviceGetHolders(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, got1 := s.getHolders()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHolders() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getHolders() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSysFsDeviceGetSlaves(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, got1 := s.getSlaves()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSlaves() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getSlaves() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSysFsDeviceGetLogicalBlockSize(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetLogicalBlockSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogicalBlockSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLogicalBlockSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetPhysicalBlockSize(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetPhysicalBlockSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPhysicalBlockSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetPhysicalBlockSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetHardwareSectorSize(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetHardwareSectorSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHardwareSectorSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetHardwareSectorSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetDriveType(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetDriveType()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDriveType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDriveType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetCapacityInBytes(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetCapacityInBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCapacityInBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCapacityInBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetDeviceType(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetDeviceType("")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDeviceType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSysFsDeviceGetDependents(t *testing.T) {
	type fields struct {
		deviceName string
		path       string
		sysPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    blockdevice.DependentBlockDevices
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Device{
				deviceName: tt.fields.deviceName,
				path:       tt.fields.path,
				sysPath:    tt.fields.sysPath,
			}
			got, err := s.GetDependents()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDependents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDependents() got = %v, want %v", got, tt.want)
			}
		})
	}
}
