/*
Copyright 2018 The OpenEBS Author

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSparseFileCreate(t *testing.T) {
	tests := map[string]struct {
		path string
		size int64
		err  bool 
	}{
		"Create a sparse file": {path: "/tmp/test.img", size: 1024, err : false},
		"Retry with same file": {path: "/tmp/test.img", size: 1024, err : false},
		"Fail to create sub-dir file" : {path: "/tmp/0/test.img", size: 1024, err : true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			createErr := SparseFileCreate(test.path, test.size)
			assert.Equal(t, test.err, createErr != nil)
		})
	}
}

func TestSparseFileDelete(t *testing.T) {
	tests := map[string]struct {
		path string
	}{
		"Delete the sparse file "      : {path: "/tmp/test.img"},
		"Retry Delete on deleted file ": {path: "/tmp/test.img"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := SparseFileDelete(test.path)
			assert.Equal(t, nil, err)
		})
	}
}


func TestSparseFileInfo(t *testing.T) {

	testFile := "/tmp/test.img"
        testFileSize := int64(1024)

	_ = SparseFileCreate( testFile, testFileSize )

	tests := map[string]struct {
		path string
		size int64
		err  bool 
	}{
		"Valid FileInfo": {path: testFile, err : false},
		"Invalid FileInfo": {path: "invalid", err : true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			info, infoErr := SparseFileInfo( test.path )
			if  infoErr == nil {
				assert.Equal(t, testFileSize, info.Size() )
			}
			assert.Equal(t, test.err, infoErr != nil)
		})
	}

	_ = SparseFileDelete( testFile )
}

