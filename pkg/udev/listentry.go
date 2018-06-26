// +build linux,cgo

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

package udev

/*
  #cgo LDFLAGS: -ludev
  #include <libudev.h>
*/
import "C"

// UdevListEntry wraps a libudev udev_list_entry object
type UdevListEntry struct {
	listEntry *C.struct_udev_list_entry
}

// newUdevListEntry a private helper function and returns a pointer to a new udev
func newUdevListEntry(ptr *C.struct_udev_list_entry) (le *UdevListEntry) {
	if ptr == nil {
		return nil
	}
	le = &UdevListEntry{
		listEntry: ptr,
	}
	return
}

// UdevListEntryGetNext return UdevListEntry struct if next device present.
// else it returns nil
func (le *UdevListEntry) UdevListEntryGetNext() *UdevListEntry {
	return newUdevListEntry(C.udev_list_entry_get_next(le.listEntry))
}

//UdevListEntryGetName return Udevice syspath.
func (le *UdevListEntry) UdevListEntryGetName() string {
	return C.GoString(C.udev_list_entry_get_name(le.listEntry))
}
