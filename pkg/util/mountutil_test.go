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

package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOsDiskPath(t *testing.T) {
	filePath := "/proc/self/mounts"
	mountPointUtil := NewMountUtil(filePath, "/")
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

func TestGetOsPartitionName(t *testing.T) {
	filePath := "/tmp/data"
	mountPoint := "/"
	fileContent1 := []byte("/dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0")
	fileContent2 := []byte("/dev/sda4 /newpath ext4 rw,relatime,errors=remount-ro,data=ordered 0 0")
	expectedErr2 := errors.New("error while geting os partition name")

	tests := map[string]struct {
		fileContent   []byte
		osPartition   string
		expectedError error
	}{
		"/ path present":     {fileContent: fileContent1, osPartition: "sda4", expectedError: nil},
		"/ path not present": {fileContent: fileContent2, osPartition: "", expectedError: expectedErr2},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := ioutil.WriteFile(filePath, test.fileContent, 0644)
			if err != nil {
				t.Fatal(err)
			}
			mountPointUtil := MountUtil{
				FilePath:   filePath,
				MountPoint: mountPoint,
			}
			partName, err := mountPointUtil.getPartitionName()
			assert.Equal(t, test.osPartition, partName)
			assert.Equal(t, test.expectedError, err)
			os.Remove(filePath)
		})
	}
	// Test case for invalid file path
	mountPointUtil := MountUtil{
		FilePath:   filePath,
		MountPoint: mountPoint,
	}
	_, err := mountPointUtil.getPartitionName()
	if err == nil {
		t.Fatal("error should not be nil for invalid path")
	}
}
