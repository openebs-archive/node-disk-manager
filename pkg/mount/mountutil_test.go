/*
Copyright 2018 OpenEBS Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMountUtil(t *testing.T) {
	filePath := "/host/proc/1/mounts"
	devPath := "/dev/sda"

	expectedMountUtil := MountUtil{
		filePath: filePath,
		devPath:  devPath,
	}

	tests := map[string]struct {
		actualMU   MountUtil
		expectedMU MountUtil
	}{
		"test for generated mount util": {newMountUtil(filePath, devPath), expectedMountUtil},
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

	tests := map[string]struct {
		devPath           string
		expectedMountAttr MountAttr
		expectedError     error
		fileContent       []byte
	}{
		"sda4 mounted at /": {
			"/dev/sda4",
			MountAttr{"/", "ext4"},
			nil,
			fileContent1,
		},
		"sda3 mounted at /home": {
			"/dev/sda3",
			MountAttr{"/home", "ext4"},
			nil,
			fileContent2,
		},
		"device is not mounted": {
			"/dev/sda3",
			MountAttr{},
			errors.New("could not get mount attributes, /dev/sda3 not mounted"),
			fileContent3,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mountUtil := newMountUtil(filePath, test.devPath)

			// create the temp file which will be read for getting attributes
			err := ioutil.WriteFile(filePath, test.fileContent, 0644)
			if err != nil {
				t.Fatal(err)
			}

			mountAttr, err := mountUtil.getMountAttr()

			assert.Equal(t, test.expectedMountAttr, mountAttr)
			assert.Equal(t, test.expectedError, err)

			// remove the temp file
			os.Remove(filePath)
		})
	}

	// invalid path tests
	mountUtil := newMountUtil(filePath, "/dev/sda3")
	_, err := mountUtil.getMountAttr()
	assert.NotNil(t, err)
}
