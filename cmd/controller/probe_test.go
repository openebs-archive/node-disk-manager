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

func (np *fakeProbe) FillDiskDetails(fakeDiskInfo *DiskInfo) {
	fakeDiskInfo.Model = fakeModel
	fakeDiskInfo.Serial = fakeSerial
	fakeDiskInfo.Vendor = fakeVendor
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
	expectedProbeList := make([]*Probe, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Probes: probes,
		Mutex:  mutex,
	}
	testProbe := &fakeProbe{}
	probe1 := &Probe{
		Priority:  2,
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	probe2 := &Probe{
		Priority:  1,
		Name:      "probe2",
		State:     true,
		Interface: testProbe,
	}
	fakeController.AddNewProbe(probe1)
	fakeController.AddNewProbe(probe2)
	expectedProbeList = append(expectedProbeList, probe2)
	expectedProbeList = append(expectedProbeList, probe1)
	tests := map[string]struct {
		actualProbeList   []*Probe
		expectedProbeList []*Probe
	}{
		"add some probes and check if they are present or not": {actualProbeList: fakeController.ListProbe(), expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

func TestSrartProbe(t *testing.T) {
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
	actualDisk := &DiskInfo{}
	expectedDisk := &DiskInfo{}
	probe1.FillDiskDetails(actualDisk)
	expectedDisk.Model = fakeModel
	expectedDisk.Serial = fakeSerial
	expectedDisk.Vendor = fakeVendor
	tests := map[string]struct {
		actualDisk   DiskInfo
		expectedDisk DiskInfo
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
	expectedDr := &DiskInfo{}
	expectedDr.Model = fakeModel
	expectedDr.Serial = fakeSerial
	expectedDr.Vendor = fakeVendor

	// create one fake Disk struct
	actualDr := &DiskInfo{}

	fakeController.FillDiskDetails(actualDr)
	tests := map[string]struct {
		actualDisk   *DiskInfo
		expectedDisk *DiskInfo
	}{
		"push resouce with 'fake-disk-uid' uuid for create resource": {actualDisk: expectedDr, expectedDisk: actualDr},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}
