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

package sysfs

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetParent(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice      *Device
		wantedDeviceName string
		wantOk           bool
	}{
		"[block] given block device is a parent": {
			sysfsDevice: &Device{
				deviceName: "sda",
				path:       "/dev/sda",
				sysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[block] given blockdevice is a partition": {
			sysfsDevice: &Device{
				deviceName: "sda4",
				path:       "/dev/sda4",
				sysPath:    "/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4/",
			},
			wantedDeviceName: "sda",
			wantOk:           true,
		},
		"[nvme] given blockdevice is a parent": {
			sysfsDevice: &Device{
				deviceName: "nvme0n1",
				path:       "/dev/nvme0n1",
				sysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/",
			},
			wantedDeviceName: "",
			wantOk:           false,
		},
		"[nvme] given blockdevice is a partition": {
			sysfsDevice: &Device{
				deviceName: "nvme0n1p1",
				path:       "/dev/nvme0n1p1",
				sysPath:    "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1/",
			},
			wantedDeviceName: "nvme0n1",
			wantOk:           true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotDeviceName, gotOk := test.sysfsDevice.getParent()
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
	tests := map[string]struct {
		fileEntries []string
		sysfsDevice *Device
		want        []string
		wantOk      bool
	}{
		"device is a partition": {
			fileEntries: nil,
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			want:   []string{},
			wantOk: true,
		},
		"device is a parent disk, with no partition": {
			fileEntries: nil,
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			want:   []string{},
			wantOk: true,
		},
		"device is a parent disk, with multiple partitions": {
			fileEntries: []string{"sdb1", "sdb2", "sdb3"},
			sysfsDevice: &Device{
				deviceName: "sdb",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdb/",
				path:       "/dev/sdb",
			},
			want:   []string{"sdb1", "sdb2", "sdb3"},
			wantOk: true,
		},
		"nvme device, with multiple partitions": {
			fileEntries: []string{"nvme0n1p1", "nvme0n1p2", "nvme0n1p3"},
			sysfsDevice: &Device{
				deviceName: "nvme0n1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/",
				path:       "/dev/nvme0n1",
			},
			want:   []string{"nvme0n1p1", "nvme0n1p2", "nvme0n1p3"},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			for _, file := range tt.fileEntries {
				os.Create(tt.sysfsDevice.sysPath + file)
			}
			got, gotOk := tt.sysfsDevice.getPartitions()
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetHolders(t *testing.T) {
	tests := map[string]struct {
		fileEntries     []string
		sysfsDevice     *Device
		createHolderDir bool
		want            []string
		wantOk          bool
	}{
		"device with empty holder directory": {
			fileEntries: []string{},
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			createHolderDir: true,
			want:            []string{},
			wantOk:          true,
		},
		"device with holders": {
			fileEntries: []string{"dm-0", "dm-1"},
			sysfsDevice: &Device{
				deviceName: "sdb",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdb/",
				path:       "/dev/sdb",
			},
			createHolderDir: true,
			want:            []string{"dm-0", "dm-1"},
			wantOk:          true,
		},
		"device without holders and no holder directory": {
			fileEntries: []string{},
			sysfsDevice: &Device{
				deviceName: "sdc",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdc/",
				path:       "/dev/sdc",
			},
			createHolderDir: false,
			want:            nil,
			wantOk:          false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createHolderDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"holders", 0700)
				for _, file := range tt.fileEntries {
					os.Create(tt.sysfsDevice.sysPath + "holders/" + file)
				}
			}
			got, gotOk := tt.sysfsDevice.getHolders()
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetSlaves(t *testing.T) {
	tests := map[string]struct {
		fileEntries    []string
		sysfsDevice    *Device
		createSlaveDir bool
		want           []string
		wantOk         bool
	}{
		"device with empty slave directory": {
			fileEntries: []string{},
			sysfsDevice: &Device{
				deviceName: "dm-0",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-0/",
				path:       "/dev/dm-0",
			},
			createSlaveDir: true,
			want:           []string{},
			wantOk:         true,
		},
		"device with slaves": {
			fileEntries: []string{"sda", "sdb"},
			sysfsDevice: &Device{
				deviceName: "dm-1",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-1/",
				path:       "/dev/dm-1",
			},
			createSlaveDir: true,
			want:           []string{"sda", "sdb"},
			wantOk:         true,
		},
		"device without slaves and no slave directory": {
			fileEntries: []string{},
			sysfsDevice: &Device{
				deviceName: "dm-2",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-2/",
				path:       "/dev/dm-2",
			},
			createSlaveDir: false,
			want:           nil,
			wantOk:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createSlaveDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"slaves", 0700)
				for _, file := range tt.fileEntries {
					os.Create(tt.sysfsDevice.sysPath + "slaves/" + file)
				}
			}
			got, gotOk := tt.sysfsDevice.getSlaves()
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetLogicalBlockSize(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice    *Device
		createQueueDir bool
		lbSize         string
		want           int64
		wantErr        bool
	}{
		"no queue directory in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			createQueueDir: false,
			lbSize:         "0",
			want:           0,
			wantErr:        true,
		},
		"valid blocksize present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			createQueueDir: true,
			lbSize:         "512",
			want:           512,
			wantErr:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createQueueDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"queue", 0700)
				file, _ := os.Create(tt.sysfsDevice.sysPath + "queue/logical_block_size")
				file.Write([]byte(tt.lbSize))
				file.Close()
			}
			got, err := tt.sysfsDevice.GetLogicalBlockSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogicalBlockSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetPhysicalBlockSize(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice    *Device
		createQueueDir bool
		pbSize         string
		want           int64
		wantErr        bool
	}{
		"no queue directory in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			createQueueDir: false,
			pbSize:         "0",
			want:           0,
			wantErr:        true,
		},
		"valid blocksize present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			createQueueDir: true,
			pbSize:         "512",
			want:           512,
			wantErr:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createQueueDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"queue", 0700)
				file, _ := os.Create(tt.sysfsDevice.sysPath + "queue/physical_block_size")
				file.Write([]byte(tt.pbSize))
				file.Close()
			}
			got, err := tt.sysfsDevice.GetPhysicalBlockSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPhysicalBlockSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetHardwareSectorSize(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice    *Device
		createQueueDir bool
		hwSize         string
		want           int64
		wantErr        bool
	}{
		"no queue directory in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			createQueueDir: false,
			hwSize:         "0",
			want:           0,
			wantErr:        true,
		},
		"valid blocksize present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			createQueueDir: true,
			hwSize:         "512",
			want:           512,
			wantErr:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createQueueDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"queue", 0700)
				file, _ := os.Create(tt.sysfsDevice.sysPath + "queue/hw_sector_size")
				file.Write([]byte(tt.hwSize))
				file.Close()
			}
			got, err := tt.sysfsDevice.GetHardwareSectorSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHardwareSectorSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetDriveType(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice    *Device
		createQueueDir bool
		rotational     string
		want           string
		wantErr        bool
	}{
		"no queue directory in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			createQueueDir: false,
			rotational:     "0",
			want:           "",
			wantErr:        true,
		},
		"valid rotational value (1) present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			createQueueDir: true,
			rotational:     "1",
			want:           blockdevice.DriveTypeHDD,
			wantErr:        false,
		},
		"valid rotational value (0) present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sdb",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdb/",
				path:       "/dev/sdb",
			},
			createQueueDir: true,
			rotational:     "0",
			want:           blockdevice.DriveTypeSSD,
			wantErr:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if tt.createQueueDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"queue", 0700)
				file, _ := os.Create(tt.sysfsDevice.sysPath + "queue/rotational")
				file.Write([]byte(tt.rotational))
				file.Close()
			}
			got, err := tt.sysfsDevice.GetDriveType()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDriveType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetCapacityInBytes(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice *Device
		size        string
		want        int64
		wantErr     bool
	}{
		"size 0 in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1/",
				path:       "/dev/sda1",
			},
			size:    "0",
			want:    0,
			wantErr: true,
		},
		"non zero size present in syspath": {
			sysfsDevice: &Device{
				deviceName: "sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
				path:       "/dev/sda",
			},
			size:    "976773168",
			want:    500107862016,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)

			file, _ := os.Create(tt.sysfsDevice.sysPath + "size")
			file.Write([]byte(tt.size))
			file.Close()

			got, err := tt.sysfsDevice.GetCapacityInBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCapacityInBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetDeviceType(t *testing.T) {
	tests := map[string]struct {
		sysfsDevice *Device
		// should be either disk / partition
		devType string

		// used for dm and md devices
		subDirectoryName string
		subFileName      string
		subFileContent   string

		want    string
		wantErr bool
	}{
		"device is a normal SCSI disk": {
			sysfsDevice: &Device{
				deviceName: "sda",
				path:       "/dev/sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "",
			subFileName:      "",
			subFileContent:   "",
			want:             blockdevice.BlockDeviceTypeDisk,
			wantErr:          false,
		},
		"device is SCSI disk partition": {
			sysfsDevice: &Device{
				deviceName: "sdb1",
				path:       "/dev/sdb1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdb/sdb1/",
			},
			devType:          blockdevice.BlockDeviceTypePartition,
			subDirectoryName: "",
			subFileName:      "",
			subFileContent:   "",
			want:             blockdevice.BlockDeviceTypePartition,
			wantErr:          false,
		},
		"device is an LVM": {
			sysfsDevice: &Device{
				deviceName: "dm-0",
				path:       "/dev/dm-0",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-0/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "dm",
			subFileName:      "uuid",
			subFileContent:   "LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt",
			want:             blockdevice.BlockDeviceTypeLVM,
			wantErr:          false,
		},
		"device is a partition on an LVM device": {
			sysfsDevice: &Device{
				deviceName: "dm-1",
				path:       "/dev/dm-1",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-1/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "dm",
			subFileName:      "uuid",
			subFileContent:   "part1-LVM-OSlVs5gIXuqSKVPukc2aGPh0AeJw31TJqYIRuRHoodYg9Jwkmyvvk0QNYK4YulHt",
			want:             blockdevice.BlockDeviceTypePartition,
			wantErr:          false,
		},
		"device is LUKS encrypted device": {
			sysfsDevice: &Device{
				deviceName: "dm-2",
				path:       "/dev/dm-2",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-2/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "dm",
			subFileName:      "uuid",
			subFileContent:   "CRYPT-LUKS1-ecc7566437ed483996273d3f50dc5871-backup",
			want:             blockdevice.BlockDeviceTypeCrypt,
			wantErr:          false,
		},
		"device is a partition on LUKS encrypted device": {
			sysfsDevice: &Device{
				deviceName: "dm-3",
				path:       "/dev/dm-3",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-3/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "dm",
			subFileName:      "uuid",
			subFileContent:   "part1-CRYPT-LUKS1-ecc7566437ed483996273d3f50dc5871-backup",
			want:             blockdevice.BlockDeviceTypePartition,
			wantErr:          false,
		},
		"device is a loop device": {
			sysfsDevice: &Device{
				deviceName: "loop7",
				path:       "/dev/loop7",
				sysPath:    "/tmp/sys/devices/virtual/block/loop7/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "",
			subFileName:      "",
			subFileContent:   "",
			want:             blockdevice.BlockDeviceTypeLoop,
			wantErr:          false,
		},
		"device is an md device (software raid)": {
			sysfsDevice: &Device{
				deviceName: "md0",
				path:       "/dev/md0",
				sysPath:    "/tmp/sys/devices/virtual/block/md0/",
			},
			devType:          blockdevice.BlockDeviceTypeDisk,
			subDirectoryName: "md",
			subFileName:      "level",
			subFileContent:   "raid0",
			want:             "raid0",
			wantErr:          false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)
			if len(tt.subDirectoryName) != 0 {
				os.MkdirAll(tt.sysfsDevice.sysPath+tt.subDirectoryName, 0700)
				f, _ := os.Create(tt.sysfsDevice.sysPath + tt.subDirectoryName + "/" + tt.subFileName)
				f.Write([]byte(tt.subFileContent))
				f.Close()
			}
			got, err := tt.sysfsDevice.GetDeviceType(tt.devType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			os.RemoveAll(tt.sysfsDevice.sysPath)
		})
	}
}

func TestSysFsDeviceGetDependents(t *testing.T) {
	tests := map[string]struct {
		partitionEntries []string
		createHolderDir  bool
		holderEntries    []string
		createSlaveDir   bool
		slaveEntries     []string
		sysfsDevice      *Device
		want             blockdevice.DependentBlockDevices
		wantErr          bool
	}{
		"parent disk with no dependents": {
			partitionEntries: []string{},
			createHolderDir:  false,
			holderEntries:    []string{},
			createSlaveDir:   false,
			slaveEntries:     []string{},
			sysfsDevice: &Device{
				deviceName: "sda",
				path:       "/dev/sda",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/",
			},
			want: blockdevice.DependentBlockDevices{
				Partitions: []string{},
				Holders:    []string{},
				Slaves:     []string{},
			},
			wantErr: false,
		},
		"parent disk with partitions": {
			partitionEntries: []string{"sdb1", "sdb2"},
			createHolderDir:  false,
			holderEntries:    []string{},
			createSlaveDir:   false,
			slaveEntries:     []string{},
			sysfsDevice: &Device{
				deviceName: "sdb",
				path:       "/dev/sdb",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdb/",
			},
			want: blockdevice.DependentBlockDevices{
				Partitions: []string{"/dev/sdb1", "/dev/sdb2"},
				Holders:    []string{},
				Slaves:     []string{},
			},
			wantErr: false,
		},
		"parent disk with holders": {
			partitionEntries: []string{},
			createHolderDir:  true,
			holderEntries:    []string{"dm-0", "dm-1"},
			createSlaveDir:   false,
			slaveEntries:     []string{},
			sysfsDevice: &Device{
				deviceName: "sdc",
				path:       "/dev/sdc",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdc/",
			},
			want: blockdevice.DependentBlockDevices{
				Partitions: []string{},
				Holders:    []string{"/dev/dm-0", "/dev/dm-1"},
				Slaves:     []string{},
			},
			wantErr: false,
		},
		"partition device without holders": {
			partitionEntries: []string{},
			createHolderDir:  false,
			holderEntries:    []string{},
			createSlaveDir:   false,
			slaveEntries:     []string{},
			sysfsDevice: &Device{
				deviceName: "sdd1",
				path:       "/dev/sdd1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sdd/sdd1",
			},
			want: blockdevice.DependentBlockDevices{
				Parent:     "/dev/sdd",
				Partitions: []string{},
				Holders:    []string{},
				Slaves:     []string{},
			},
			wantErr: false,
		},
		"partition device with holders": {
			partitionEntries: []string{},
			createHolderDir:  true,
			holderEntries:    []string{"dm-0"},
			createSlaveDir:   false,
			slaveEntries:     []string{},
			sysfsDevice: &Device{
				deviceName: "sde1",
				path:       "/dev/sde1",
				sysPath:    "/tmp/sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sde/sde1/",
			},
			want: blockdevice.DependentBlockDevices{
				Parent:     "/dev/sde",
				Partitions: []string{},
				Holders:    []string{"/dev/dm-0"},
				Slaves:     []string{},
			},
			wantErr: false,
		},
		"device with both slaves and holders": {
			partitionEntries: []string{},
			createHolderDir:  true,
			holderEntries:    []string{"dm-1", "dm-2"},
			createSlaveDir:   true,
			slaveEntries:     []string{"sdb", "sdc"},
			sysfsDevice: &Device{
				deviceName: "dm-0",
				path:       "/dev/dm-0",
				sysPath:    "/tmp/sys/devices/virtual/block/dm-0/",
			},
			want: blockdevice.DependentBlockDevices{
				Partitions: []string{},
				Holders:    []string{"/dev/dm-1", "/dev/dm-2"},
				Slaves:     []string{"/dev/sdb", "/dev/sdc"},
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			os.MkdirAll(tt.sysfsDevice.sysPath, 0700)

			// partition entries
			for _, file := range tt.partitionEntries {
				os.Create(tt.sysfsDevice.sysPath + file)
			}

			// holder entries
			if tt.createHolderDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"holders", 0700)
				for _, file := range tt.holderEntries {
					os.Create(tt.sysfsDevice.sysPath + "holders/" + file)
				}
			}

			// slave entries
			if tt.createSlaveDir {
				os.MkdirAll(tt.sysfsDevice.sysPath+"slaves", 0700)
				for _, file := range tt.slaveEntries {
					os.Create(tt.sysfsDevice.sysPath + "slaves/" + file)
				}
			}

			got, err := tt.sysfsDevice.GetDependents()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDependents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
