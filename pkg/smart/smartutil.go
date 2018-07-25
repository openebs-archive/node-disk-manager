/*
Copyright 2018 The OpenEBS Authors.

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

package smart

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

//intSize is the size in bytes (converted to integer) of 0
const intSize int = int(unsafe.Sizeof(0))

// A ByteOrder specifies how to convert byte sequences
// into 16-, 32-, or 64-bit unsigned integers.
var (
	NativeEndian binary.ByteOrder
)

// init determines native endianness of a system
func init() {
	i := 0x1
	b := (*[intSize]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		// LittleEndian is the little-endian implementation of ByteOrde
		NativeEndian = binary.LittleEndian
	} else {
		// BigEndian is the Big-endian implementation of ByteOrde
		NativeEndian = binary.BigEndian
	}
}

// ErrorCollector Struct is a struct for map of errors
type ErrorCollector struct {
	errors map[string]error
}

// NewErrorCollector returns a pointer to the ErrorCollector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{errors: map[string]error{}}
}

// Collect function is used to collect errors corresponding to the keys given to it
func (c *ErrorCollector) Collect(key string, e error) bool {
	if e != nil {
		c.errors[key] = e
		return true
	}
	return false
}

// Error is used to return all the collected errors as a map
func (c *ErrorCollector) Error() (errorMap map[string]error) {
	return c.errors
}

// MSignificantBit finds the most significant bit set in a uint
func MSignificantBit(i uint) int {
	if i == 0 {
		return 0
	}
	return bits.Len(i) - 1
}
