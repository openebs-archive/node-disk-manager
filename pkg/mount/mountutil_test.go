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
