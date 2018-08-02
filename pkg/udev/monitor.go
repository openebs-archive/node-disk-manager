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

import (
	"errors"
)

// UdevMonitor wraps a libudev monitor device object
type UdevMonitor struct {
	umptr *C.struct_udev_monitor
}

// newUdevMonitor is a helper function and returns a pointer to a new monitor device.
// The argument ptr is a pointer to the underlying C udev_device_monitor structure.
// The function returns nil if the pointer passed is NULL.
func newUdeviceMonitor(ptr *C.struct_udev_monitor) (*UdevMonitor, error) {
	// If passed a NULL pointer, return nil
	if ptr == nil {
		return nil, errors.New("unable to create Udevenumerate object for for null struct struct_udev_monitor")
	}
	// Create a new monitor device object
	um := &UdevMonitor{
		umptr: ptr,
	}
	// Return the monitor device object
	return um, nil
}

// AddSubsystemFilter adds subsystem filter in UdevMonitor struct.
// This filter is efficiently executed inside kernel, and libudev
// subscribers will usually not be woken up for devices which do not match.
func (um *UdevMonitor) AddSubsystemFilter(key string) error {
	subsystem := C.CString(key)
	defer freeCharPtr(subsystem)
	ret := C.udev_monitor_filter_add_match_subsystem_devtype(um.umptr, subsystem, nil)
	if ret < 0 {
		return errors.New("unable to apply filter")
	}
	return nil
}

// EnableReceiving binds udev_monitor socket to event source.
func (um *UdevMonitor) EnableReceiving() error {
	ret := C.udev_monitor_enable_receiving(um.umptr)
	if ret < 0 {
		return errors.New("unable to enable receiving udev")
	}
	return nil
}

// GetFd retrieves socket file descriptor associated with monitor.
func (um *UdevMonitor) GetFd() (int, error) {
	ret := int(C.udev_monitor_get_fd(um.umptr))
	if ret < 0 {
		return ret, errors.New("unable to get fd from monitor")
	}
	return ret, nil
}

// ReceiveDevice receives data from udev monitor socket, allocate a
// new udev device, fill in received data, and return Udevice struct.
func (um *UdevMonitor) ReceiveDevice() (*UdevDevice, error) {
	return newUdevDevice(C.udev_monitor_receive_device(um.umptr))
}

// UdevMonitorUnref frees udev monitor structure.
func (um *UdevMonitor) UdevMonitorUnref() {
	C.udev_monitor_unref(um.umptr)
}
