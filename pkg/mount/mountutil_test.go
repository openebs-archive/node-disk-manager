/*
Copyright 2019 OpenEBS Authors.

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

package mount

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMountUtil(t *testing.T) {
	filePath := "/host/proc/1/mounts"
	devPath := "/dev/sda"
	mountPoint := "/home"
	// TODO
	expectedMountUtil1 := DiskMountUtil{
		filePath: filePath,
		devPath:  devPath,
	}
	expectedMountUtil2 := DiskMountUtil{
		filePath:   filePath,
		mountPoint: mountPoint,
	}

	tests := map[string]struct {
		actualMU   DiskMountUtil
		expectedMU DiskMountUtil
	}{
		"test for generated mount util with devPath":    {NewMountUtil(filePath, devPath, ""), expectedMountUtil1},
		"test for generated mount util with mountpoint": {NewMountUtil(filePath, "", mountPoint), expectedMountUtil2},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedMU, test.actualMU)
		})
	}
}

func TestGetMountAttr(t *testing.T) {
	filePath := "/tmp/data"
	fileContent1 := []byte("/dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0")
	fileContent2 := []byte("/dev/sda3 /home ext4 rw,relatime,errors=remount-ro,data=ordered 0 0")
	fileContent3 := []byte("sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0")

	mountAttrTests := map[string]struct {
		devPath           string
		expectedMountAttr DeviceMountAttr
		expectedError     error
		fileContent       []byte
	}{
		"sda4 mounted at /": {
			"/dev/sda4",
			DeviceMountAttr{MountPoint: "/", FileSystem: "ext4"},
			nil,
			fileContent1,
		},
		"sda3 mounted at /home": {
			"/dev/sda3",
			DeviceMountAttr{MountPoint: "/home", FileSystem: "ext4"},
			nil,
			fileContent2,
		},
		"device is not mounted": {
			"/dev/sda3",
			DeviceMountAttr{},
			errors.New("could not get device mount attributes, Path/MountPoint not present in mounts file"),
			fileContent3,
		},
	}
	for name, test := range mountAttrTests {
		t.Run(name, func(t *testing.T) {
			mountUtil := NewMountUtil(filePath, test.devPath, "")

			// create the temp file which will be read for getting attributes
			err := ioutil.WriteFile(filePath, test.fileContent, 0644)
			if err != nil {
				t.Fatal(err)
			}

			mountAttr, err := mountUtil.getDeviceMountAttr(mountUtil.getMountName)

			assert.Equal(t, test.expectedMountAttr, mountAttr)
			assert.Equal(t, test.expectedError, err)

			// remove the temp file
			os.Remove(filePath)
		})
	}
	// TODO tests that use mountUtil.getPartitionName in getDeviceMountAttr

	// invalid path mountAttrTests
	mountUtil := NewMountUtil(filePath, "/dev/sda3", "")
	_, err := mountUtil.getDeviceMountAttr(mountUtil.getMountName)
	assert.NotNil(t, err)
}

func TestGetPartitionName(t *testing.T) {
	mountLine := "/dev/sda4 /home ext4 rw,relatime,errors=remount-ro,data=ordered 0 0"
	mountPoint1 := "/home"
	mountPoint2 := "/"
	tests := map[string]struct {
		expectedAttr DeviceMountAttr
		expectedOk   bool
		mountPoint   string
		line         string
	}{
		"mount point is present in line":     {DeviceMountAttr{DevPath: "sda4"}, true, mountPoint1, mountLine},
		"mount point is not present in line": {DeviceMountAttr{}, false, mountPoint2, mountLine},
		"mountline is empty":                 {DeviceMountAttr{}, false, mountPoint2, ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mountPointUtil := NewMountUtil("", "", test.mountPoint)
			mountAttr, ok := mountPointUtil.getPartitionName(test.line)
			assert.Equal(t, test.expectedAttr, mountAttr)
			assert.Equal(t, test.expectedOk, ok)
		})
	}
}

func TestGetMountName(t *testing.T) {
	mountLine := "/dev/sda4 /home ext4 rw,relatime,errors=remount-ro,data=ordered 0 0"
	devPath1 := "/dev/sda4"
	devPath2 := "/dev/sda3"
	fsType := "ext4"
	mountPoint := "/home"
	tests := map[string]struct {
		expectedMountAttr DeviceMountAttr
		expectedOk        bool
		devPath           string
		line              string
	}{
		"device sda4 is mounted":     {DeviceMountAttr{MountPoint: mountPoint, FileSystem: fsType}, true, devPath1, mountLine},
		"device sda3 is not mounted": {DeviceMountAttr{}, false, devPath2, mountLine},
		"mount line is empty":        {DeviceMountAttr{}, false, devPath2, ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mountPointUtil := NewMountUtil("", test.devPath, "")
			attr, ok := mountPointUtil.getMountName(test.line)
			assert.Equal(t, test.expectedMountAttr, attr)
			assert.Equal(t, test.expectedOk, ok)
		})
	}
}

func TestOsDiskPath(t *testing.T) {
	filePath := "/proc/self/mounts"
	mountPointUtil := NewMountUtil(filePath, "", "/")
	path, err := mountPointUtil.GetDiskPath()
	tests := map[string]struct {
		actualPath    string
		actualError   error
		expectedError error
	}{
		"test case for os disk path": {actualPath: path, actualError: err, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := filepath.EvalSymlinks(test.actualPath)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestGetParentBlockDevice(t *testing.T) {
	tests := map[string]struct {
		syspath                   string
		expectedParentBlockDevice string
		expectedOk                bool
	}{
		"getting parent of a main blockdevice itself": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0d.0/ata1/host0/target0:0:0/0:0:0:0/block/sda",
			expectedParentBlockDevice: "sda",
			expectedOk:                true,
		},
		"getting parent of a partition": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0d.0/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda1",
			expectedParentBlockDevice: "sda",
			expectedOk:                true,
		},
		"getting parent of main NVMe blockdevice": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1",
			expectedParentBlockDevice: "nvme0n1",
			expectedOk:                true,
		},
		"getting parent of partitioned NVMe blockdevice": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1",
			expectedParentBlockDevice: "nvme0n1",
			expectedOk:                true,
		},
		"getting parent of wrong disk": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0",
			expectedParentBlockDevice: "",
			expectedOk:                false,
		},
		"giving a wrong syspath": {
			syspath:                   "/sys/devices/pci0000:00/0000:00:0e.0",
			expectedParentBlockDevice: "",
			expectedOk:                false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parentBlockDevice, ok := getParentBlockDevice(test.syspath)
			assert.Equal(t, test.expectedParentBlockDevice, parentBlockDevice)
			assert.Equal(t, test.expectedOk, ok)
		})
	}
}

func TestParseRootDeviceLink(t *testing.T) {
	tests := map[string]struct {
		content       string
		expectedPath  string
		expectedError error
	}{
		"empty content": {
			"",
			"",
			ErrCouldNotFindRootDevice,
		},
		"single line with root only": {
			"root=UUID=d41162ba-25e4-4c44-8793-2abef96d27e9",
			"/dev/disk/by-uuid/d41162ba-25e4-4c44-8793-2abef96d27e9",
			nil,
		},
		"single line with multiple attributes": {
			"BOOT_IMAGE=/boot/vmlinuz-5.4.0-48-generic root=UUID=d41162ba-25e4-4c44-8793-2abef96d27e9 ro intel_iommu=on quiet splash vt.handoff=7",
			"/dev/disk/by-uuid/d41162ba-25e4-4c44-8793-2abef96d27e9",
			nil,
		},
		"single line without root attribute": {
			"BOOT_IMAGE=/boot/vmlinuz-5.4.0-48-generic ro intel_iommu=on quiet splash vt.handoff=7",
			"",
			ErrCouldNotFindRootDevice,
		},
		"multi line with multiple attributes": {
			"\n\nBOOT_IMAGE=/boot/vmlinuz-5.4.0-48-generic root=PARTUUID=325c5bfa-08a8-433c-bc62-2dd5255213fd ro\n",
			"/dev/disk/by-partuuid/325c5bfa-08a8-433c-bc62-2dd5255213fd",
			nil,
		},
		"single line with root on dm device, (simulates cmdline in GKE)": {
			"BOOT_IMAGE=/syslinux/vmlinuz.A root=/dev/dm-0",
			"/dev/dm-0",
			nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			actualPath, actualError := parseRootDeviceLink(strings.NewReader(test.content))

			assert.Equal(t, test.expectedError, actualError)

			if actualError == nil {
				assert.Equal(t, test.expectedPath, actualPath)
			}
		})
	}
}

func TestGetSoftLinkForPartition(t *testing.T) {
	tests := map[string]string{
		"sda1":    "/sys/class/block/sda1",
		"nvme0n1": "/sys/class/block/nvme0n1",
	}

	for partition, expectedSoftlink := range tests {
		t.Run(partition, func(t *testing.T) {
			actualSoftLink, actualError := getSoftLinkForPartition(partition)
			assert.NoError(t, actualError)
			assert.Equal(t, expectedSoftlink, actualSoftLink)
		})
	}

	t.Run("root", func(t *testing.T) {
		actualSoftLink, actualError := getSoftLinkForPartition("root")
		assert.NoError(t, actualError)
		assert.NotEqual(t, "/sys/class/block/root", actualSoftLink)
		assert.True(t, strings.HasPrefix(actualSoftLink, "/sys/class/block/"))
	})
}

func TestGetDiskDevPath_WithRoot(t *testing.T) {
	path, err := getPartitionDevPath("root")

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(path, "/dev/"))
}
