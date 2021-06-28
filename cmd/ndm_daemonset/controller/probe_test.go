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

package controller

import (
	"sync"
	"testing"
	"time"

	bd "github.com/openebs/node-disk-manager/blockdevice"

	"github.com/stretchr/testify/assert"
)

const (
	fakeModel  = "fake-model-number"
	fakeSerial = "fake-serial-number"
	fakeVendor = "fake-vendor"
)

var messageChannel = make(chan string)
var message = "This is a message from start method"

type fakeProbe struct{}

func (np *fakeProbe) Start() {
	messageChannel <- message
}

func (np *fakeProbe) FillBlockDeviceDetails(fakeBlockDevice *bd.BlockDevice) {
	fakeBlockDevice.DeviceAttributes.Model = fakeModel
	fakeBlockDevice.DeviceAttributes.Serial = fakeSerial
	fakeBlockDevice.DeviceAttributes.Vendor = fakeVendor
}

//Add one new probe and get the list of the probes and match them
func TestAddNewProbe(t *testing.T) {
	probes := make([]*Probe, 0)
	expectedProbeList := make([]*Probe, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Probes: probes,
		Mutex:  mutex,
	}
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	fakeController.AddNewProbe(probe1)
	expectedProbeList = append(expectedProbeList, probe1)
	tests := map[string]struct {
		actualProbeList   []*Probe
		expectedProbeList []*Probe
	}{
		"add one probe and check if it is present or not": {actualProbeList: fakeController.Probes, expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

//Add some new probes and get the list of the probes and match them
func TestListProbe(t *testing.T) {
	probes := make([]*Probe, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Probes: probes,
		Mutex:  mutex,
	}
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Priority:  3,
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	probe2 := &Probe{
		Priority:  2,
		Name:      "probe2",
		State:     true,
		Interface: testProbe,
	}
	probe3 := &Probe{
		Priority:  4,
		Name:      "probe3",
		State:     false,
		Interface: testProbe,
	}
	probe4 := &Probe{
		Priority:  1,
		Name:      "probe4",
		State:     true,
		Interface: testProbe,
	}

	fakeController.AddNewProbe(probe1)
	fakeController.AddNewProbe(probe2)
	fakeController.AddNewProbe(probe3)
	fakeController.AddNewProbe(probe4)

	tests := map[string]struct {
		actualProbeList   []*Probe
		expectedProbeList []*Probe
	}{
		"list all enabled probes": {
			actualProbeList:   fakeController.ListProbe(),
			expectedProbeList: []*Probe{probe4, probe2, probe1},
		},
		"list selective probes, all required probes enabled": {
			actualProbeList:   fakeController.ListProbe("probe1", "probe2"),
			expectedProbeList: []*Probe{probe2, probe1},
		},
		"list selective probes, some disabled": {
			actualProbeList:   fakeController.ListProbe("probe2", "probe4", "probe3"),
			expectedProbeList: []*Probe{probe4, probe2},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

func TestStartProbe(t *testing.T) {
	var msg1 string
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	go probe1.Start()
	select {
	case res := <-messageChannel:
		msg1 = res
	case <-time.After(1 * time.Second):
		msg1 = ""
	}

	tests := map[string]struct {
		actualMessage   string
		expectedMessage string
	}{
		"comparing message from start method": {actualMessage: msg1, expectedMessage: message},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedMessage, test.actualMessage)
		})
	}
}

func TestFillDiskDetails(t *testing.T) {
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	actualDisk := &bd.BlockDevice{}
	expectedDisk := &bd.BlockDevice{}
	probe1.FillBlockDeviceDetails(actualDisk)
	expectedDisk.DeviceAttributes.Model = fakeModel
	expectedDisk.DeviceAttributes.Serial = fakeSerial
	expectedDisk.DeviceAttributes.Vendor = fakeVendor
	tests := map[string]struct {
		actualDisk   bd.BlockDevice
		expectedDisk bd.BlockDevice
	}{
		"comparing diskinfo struct after feeling details": {actualDisk: *actualDisk, expectedDisk: *expectedDisk},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}

func TestFillDetails(t *testing.T) {
	probes := make([]*Probe, 0)
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	probes = append(probes, probe1)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Probes: probes,
		Mutex:  mutex,
	}

	// create one fake Disk struct
	expectedDr := &bd.BlockDevice{}
	expectedDr.DeviceAttributes.Model = fakeModel
	expectedDr.DeviceAttributes.Serial = fakeSerial
	expectedDr.DeviceAttributes.Vendor = fakeVendor

	// create one fake Disk struct
	actualDr := &bd.BlockDevice{}

	fakeController.FillBlockDeviceDetails(actualDr)
	tests := map[string]struct {
		actualDisk   *bd.BlockDevice
		expectedDisk *bd.BlockDevice
	}{
		"push resource with 'fake-disk-uid' uuid for create resource": {actualDisk: actualDr, expectedDisk: expectedDr},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}
