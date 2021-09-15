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

// +build linux, cgo

package blkid

/*
 #cgo LDFLAGS: -lblkid
#include "blkid/blkid.h"
#include "string.h"
#include "stdlib.h"
*/
import "C"
import (
	"unsafe"
)

const (
	fsTypeIdentifier = "TYPE"
)

type DeviceIdentifier struct {
	DevPath string
}

// GetOnDiskFileSystem returns the filesystem present on the disk by reading from the disk
// using libblkid
func (di *DeviceIdentifier) GetOnDiskFileSystem() string {
	var blkidType *C.char
	blkidType = C.CString(fsTypeIdentifier)
	defer C.free(unsafe.Pointer(blkidType))

	var device *C.char
	device = C.CString(di.DevPath)
	defer C.free(unsafe.Pointer(device))

	var fstype *C.char
	fstype = C.blkid_get_tag_value(nil, blkidType, device)
	defer C.free(unsafe.Pointer(fstype))

	return C.GoString(fstype)
}
