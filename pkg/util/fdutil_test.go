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
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFD_SET(t *testing.T) {
	tests := map[string]struct {
		x        int
		expected syscall.FdSet
	}{
		"fd set test for fd value 0": {x: 0, expected: syscall.FdSet{[32]int32{1}}},
		"fd set test for fd value 1": {x: 1, expected: syscall.FdSet{[32]int32{2}}},
		"fd set test for fd value 7": {x: 7, expected: syscall.FdSet{[32]int32{128}}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := &syscall.FdSet{}
			FD_SET(result, test.x)
			assert.Equal(t, test.expected, *result)
		})
	}
}

func TestFD_ISSET(t *testing.T) {
	tests := map[string]struct {
		x        int
		fdset    syscall.FdSet
		expected bool
	}{
		"set 0 and check with 2^0": {x: 0, fdset: syscall.FdSet{[32]int32{1}}, expected: true},
		"set 1 and check with 2^1": {x: 1, fdset: syscall.FdSet{[32]int32{2}}, expected: true},
		"set 7 and check with 2^7": {x: 7, fdset: syscall.FdSet{[32]int32{128}}, expected: true},
		"set 4 and check with 6":   {x: 4, fdset: syscall.FdSet{[32]int32{6}}, expected: false},
		"set 2 and check with 2":   {x: 2, fdset: syscall.FdSet{[32]int32{2}}, expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := FD_ISSET(&test.fdset, test.x)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFD_ZERO(t *testing.T) {
	tests := map[string]struct {
		x        int
		expected syscall.FdSet
	}{
		"fd zero test for fdset value {[16]int64{1}}":   {x: 0, expected: syscall.FdSet{[32]int32{1}}},
		"fd zero test for fdset value {[16]int64{2}}":   {x: 1, expected: syscall.FdSet{[32]int32{2}}},
		"fd zero test for fdset value {[16]int64{128}}": {x: 7, expected: syscall.FdSet{[32]int32{128}}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := &syscall.FdSet{}
			FD_ZERO(&test.expected)
			assert.Equal(t, test.expected, *result)
		})
	}
}
