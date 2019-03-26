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

package controller

import (
	//	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestGetSparseFileDir verifies that a valid sparse file directory
//  is returned as per the environment variable ( SPARSE_FILE_PATH )
func TestGetSparseFileDir(t *testing.T) {

	tests := map[string]struct {
		envSparseDir string
		expectedPath string
	}{
		"When path is not set":     {envSparseDir: "", expectedPath: ""},
		"When invalid path is set": {envSparseDir: "invalid", expectedPath: ""},
		"When valid path is set":   {envSparseDir: "/tmp", expectedPath: "/tmp"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(EnvSparseFileDir, test.envSparseDir)
			assert.Equal(t, test.expectedPath, GetSparseFileDir())
		})
	}

}

// TestGetSparseFileCount verifies that a valid sparse file count
//  is returned as per the environment variable ( SPARSE_FILE_COUNT ).
//  If no value is set, default count is returned.
func TestGetSparseFileCount(t *testing.T) {

	defaultCount, econv := strconv.Atoi(SparseFileDefaultCount)
	if econv != nil {
		defaultCount = 0
	}

	tests := map[string]struct {
		envFileCount  string
		expectedCount int
	}{
		"When count is not set":     {envFileCount: "", expectedCount: defaultCount},
		"When valid count is set":   {envFileCount: "2", expectedCount: 2},
		"When invalid count is set": {envFileCount: "z", expectedCount: 0},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(EnvSparseFileCount, test.envFileCount)
			assert.Equal(t, test.expectedCount, GetSparseFileCount())
		})
	}

}

// TestGetSparseFileSize verifies that a valid sparse file size
//  is returned as per the environment variable ( SPARSE_FILE_SIZE ).
//  If no value is set, default size is returned.
func TestGetSparseFileSize(t *testing.T) {

	defaultSize := SparseFileDefaultSize
	minSize := SparseFileMinSize

	tests := map[string]struct {
		envFileSize  string
		expectedSize int64
	}{
		"When size is not set":                {envFileSize: "", expectedSize: defaultSize},
		"When valid size is set":              {envFileSize: "2000000000", expectedSize: int64(2000000000)},
		"When file size is given as exponent": {envFileSize: "1.073741824e+11", expectedSize: int64(107374182400)},
		"When less than min size is set":      {envFileSize: "100", expectedSize: minSize},
		"When invalid size is set":            {envFileSize: "z", expectedSize: 0},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(EnvSparseFileSize, test.envFileSize)
			assert.Equal(t, test.expectedSize, GetSparseFileSize())
		})
	}

}

// TestCheckAndCreateSparseFile verifies that a sparse file is created
//  only when it doesn't already exist.
func TestCheckAndCreateSparseFile(t *testing.T) {

	testFile := "/tmp/test.img"
	testFileSize := int64(1000)
	testFileSize1 := int64(2000)

	tests := map[string]struct {
		fileName     string
		fileSize     int64
		expectedSize int64
	}{
		"Create New File": {
			fileName:     testFile,
			fileSize:     testFileSize,
			expectedSize: testFileSize,
		},
		"Reuse Existing File": {
			fileName:     testFile,
			fileSize:     testFileSize1,
			expectedSize: testFileSize,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := CheckAndCreateSparseFile(test.fileName, test.fileSize)
			assert.Equal(t, nil, err)
			if err != nil {
				aFileInfo, errF := util.SparseFileInfo(test.fileName)
				assert.Equal(t, nil, errF)
				if errF != nil {
					assert.Equal(t, test.expectedSize, aFileInfo.Size())
				}
			}
		})
	}

	util.SparseFileDelete(testFile)
}

func TestGetActiveSparseDisksUuids(t *testing.T) {
	if _, err := os.Create("/tmp/0-ndm-sparse.img"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("/tmp/0-ndm-sparse.img")
	if _, err := os.Create("/tmp/1-ndm-sparse.img"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("/tmp/1-ndm-sparse.img")
	sparseUids := make([]string, 0)
	sparseUids = append(sparseUids, "sparse-2b3468d4b928c7e048ad8747ba710e4c")
	sparseUids = append(sparseUids, "sparse-af2cd3d402e3447e315aadb7e7b46a34")
	tests := map[string]struct {
		expectedSparseDiskUuids []string
		sparseFileDir           string
	}{
		"When dir is valid":                  {sparseFileDir: "/tmp", expectedSparseDiskUuids: sparseUids},
		"When dir is invalid or not present": {sparseFileDir: "/invalid", expectedSparseDiskUuids: make([]string, 0)},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv(EnvSparseFileDir, test.sparseFileDir)
			assert.Equal(t, test.expectedSparseDiskUuids, GetActiveSparseDisksUuids("instance-1"))
			os.Unsetenv(EnvSparseFileDir)
		})
	}
}
