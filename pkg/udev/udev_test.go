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

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUdev(t *testing.T) {
	// Creating one udev struct with nil pointer it should return an error.
	_, err := newUdev(nil)
	assert.Equal(t, errors.New("unable to create Udevice object for null struct struct_udev"), err)

	// Creating one udev struct with struct_udev mock data and compare both struct.
	sUdev := mockDataStructUdev()
	uDev2, err := newUdev(sUdev)
	rxpecteddev1 := Udev{
		uptr: sUdev,
	}
	assert.Equal(t, rxpecteddev1, *uDev2)
	assert.Equal(t, nil, err)
	defer uDev2.UnrefUdev()

	// Creating one udev object for successful creation error should be nil.
	uDev3, err := NewUdev()
	if uDev3 != nil {
		assert.Equal(t, nil, err)
		defer uDev3.UnrefUdev()
	}
}

func TestNewUdevEnumerate(t *testing.T) {
	// In testcase1 enumerate one udev struct which is not nil. It should not return error.
	// In testcase2 enumerate one udev struct which is nil. It should return error.
	dev1, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer dev1.UnrefUdev()
	devenu1, err := dev1.NewUdevEnumerate()
	if err != nil {
		t.Fatal(err)
	}
	devenu1.UnrefUdevEnumerate()
	dev2, _ := newUdev(nil)
	if dev2 != nil {
		t.Fatal("udev object should be nil for null pointer")
	}
	_, err = dev2.NewUdevEnumerate()
	assert.Equal(t, errors.New("unable to create Udevenumerate object for for null struct struct_udev_enumerate"), err)
}

func TestNewDeviceFromSysPath(t *testing.T) {
	// For invalid syspath NewDeviceFromSysPath() should return an error
	udev, err := NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	defer udev.UnrefUdev()
	device1, err := udev.NewDeviceFromSysPath("")
	if device1 != nil {
		t.Fatal("device should be nil for invalid syspath")
	}
	assert.Equal(t, errors.New("unable to create Udevice object for for null struct struct_udev_device"), err)
	diskDetails, err := MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	device2, _ := udev.NewDeviceFromSysPath(diskDetails.SysPath)
	if device2 != nil {
		assert.Equal(t, diskDetails.SysPath, device2.GetSyspath())
		defer device2.UdevDeviceUnref()
	}
}
