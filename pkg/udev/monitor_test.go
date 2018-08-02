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

func TestNewDeviceFromNetlink(t *testing.T) {
	udev, err := NewUdev()
	if err != nil {
		t.Error(err)
	}
	defer udev.UnrefUdev()
	// For valid monitor source NewUdeviceMon() should return monitor device pointer and nil error.
	udevMonitor, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	defer udevMonitor.UdevMonitorUnref()
	// For invalid monitor source NewUdeviceMon() should return nil monitor device pointer and error.
	_, err = udev.NewDeviceFromNetlink("block")
	if err != nil {
		assert.Equal(t, errors.New("unable to create Udevenumerate object for for null struct struct_udev_monitor"), err)
	}
}

func TestAddFilterSubsystem(t *testing.T) {
	// for successful AddFilterSubsystem() it should return nil error.
	udev, err := NewUdev()
	if err != nil {
		t.Error(err)
	}
	defer udev.UnrefUdev()
	udevMonitor, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	defer udevMonitor.UdevMonitorUnref()
	assert.Equal(t, nil, udevMonitor.AddSubsystemFilter("block"))
}

func TestEnableMonitorReceiving(t *testing.T) {
	// for successful EnableMonitorReceiving() it should return nil error.
	// if we free udevMonitor pointer then it should return an error.
	udev, err := NewUdev()
	if err != nil {
		t.Error(err)
	}
	defer udev.UnrefUdev()
	udevMonitor, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	defer udevMonitor.UdevMonitorUnref()
	err = udevMonitor.EnableReceiving()
	if err != nil {
		t.Error("expected nil error")
	}
	udevMonitor1, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	udevMonitor1.UdevMonitorUnref()
	err = udevMonitor1.EnableReceiving()
	assert.Equal(t, errors.New("unable to enable receiving udev"), err)

}

func TestGetFd(t *testing.T) {
	// for successful TestGetFd() it should return nil error
	// and fd value should be greater than 0.
	udev, err := NewUdev()
	if err != nil {
		t.Error(err)
	}
	defer udev.UnrefUdev()
	udevMonitor, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	defer udevMonitor.UdevMonitorUnref()
	fd, err := udevMonitor.GetFd()
	if fd <= 0 {
		t.Error("fd value should be greater than 0")
	}
	assert.Equal(t, nil, err)
}

func TestUdevMonitorNewDevice(t *testing.T) {
	udev, err := NewUdev()
	if err != nil {
		t.Error(err)
	}
	defer udev.UnrefUdev()
	udevMonitor, err := udev.NewDeviceFromNetlink("udev")
	if err != nil {
		t.Fatal(err)
	}
	defer udevMonitor.UdevMonitorUnref()
	_, err = udevMonitor.ReceiveDevice()
	if err != nil {
		assert.Equal(t, errors.New("unable to create Udevice object for for null struct struct_udev_device"), err)
	}
}
