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
	}{
		"Create a sparse file of 1G - test-1g.img": {path: "/tmp/test-1g.img", size: 1073741824},
		"Retry with same file":                     {path: "/tmp/test-1g.img", size: 1073741824},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := SparseFileCreate(test.path, test.size)
			assert.Equal(t, nil, err)
		})
	}
}

func TestSparseFileDelete(t *testing.T) {
	tests := map[string]struct {
		path string
	}{
		"Delete the sparse file of 1G -test-1g.img":       {path: "/tmp/test-1g.img"},
		"Retry Delete on deleted file of 1G -test-1g.img": {path: "/tmp/test-1g.img"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := SparseFileDelete(test.path)
			assert.Equal(t, nil, err)
		})
	}
}
