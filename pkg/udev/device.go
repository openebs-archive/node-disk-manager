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
import "errors"

// UdevDevice wraps a libudev device object
type UdevDevice struct {
	udptr *C.struct_udev_device
}

// newUdevDevice is a private helper function and returns a pointer to a new device
// it returns nil if the pointer passed is NULL.
func newUdevDevice(ptr *C.struct_udev_device) (*UdevDevice, error) {
	if ptr == nil {
		return nil, errors.New("unable to create Udevice object for for null struct struct_udev_device")
	}
	ud := &UdevDevice{
		udptr: ptr,
	}
	return ud, nil
}

// GetPropertyValue retrieves the value of a device property
func (ud *UdevDevice) GetPropertyValue(key string) string {
	k := C.CString(key)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_property_value(ud.udptr, k))
}

// GetSysattrValue retrieves the content of a sys attribute file
// returns an empty string if there is no sys attribute value.
// The retrieved value is cached in the device. Repeated calls
// will return the same value and not open the attribute again.
func (ud *UdevDevice) GetSysattrValue(sysattr string) string {
	k := C.CString(sysattr)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_sysattr_value(ud.udptr, k))
}

// GetDevtype returns the devtype string of the udev device.
func (ud *UdevDevice) GetDevtype() string {
	return C.GoString(C.udev_device_get_devtype(ud.udptr))
}

// GetDevnode returns the device node file name belonging to the udev device.
// The path is an absolute path, and starts with the device directory.
func (ud *UdevDevice) GetDevnode() string {
	return C.GoString(C.udev_device_get_devnode(ud.udptr))
}

// GetAction returns device action when it is monitored.
// It can be add,remove,online,offline,change
func (ud *UdevDevice) GetAction() string {
	return C.GoString(C.udev_device_get_action(ud.udptr))
}

// UdevDeviceUnref frees udev_device structure.
func (ud *UdevDevice) UdevDeviceUnref() {
	C.udev_device_unref(ud.udptr)
}
