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

package udevevent

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	event := newEvent()
	if event == nil {
		t.Error("event pointer should not be nil")
	}
}

func TestProcess(t *testing.T) {
	actualEvent := newEvent()
	osDiskDetails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	udev, err := libudevwrapper.NewUdev()
	if err != nil {
		t.Fatal(err)
	}
	device, err := udev.NewDeviceFromSysPath(osDiskDetails.SysPath)
	if err != nil {
		t.Fatal(err)
	}
	actualEvent.process(device)
	// this is done so that the legacy UUID is not used.
	// TODO. can be removed once legacy UUID is completely removed.
	actualEvent.eventDetails.Devices[0].UUID = ""

	// creating mock event
	expectedEvent := newEvent()
	diskInfo := make([]*blockdevice.BlockDevice, 0)
	deviceDetails := &blockdevice.BlockDevice{}
	deviceDetails.SysPath = osDiskDetails.SysPath
	deviceDetails.DevPath = osDiskDetails.DevNode

	deviceDetails.DeviceAttributes.DeviceType = osDiskDetails.DevType
	deviceDetails.DeviceAttributes.WWN = osDiskDetails.Wwn
	deviceDetails.DeviceAttributes.Serial = osDiskDetails.Serial
	deviceDetails.DeviceAttributes.Model = osDiskDetails.Model
	deviceDetails.DeviceAttributes.Vendor = osDiskDetails.Vendor
	deviceDetails.DeviceAttributes.IDType = osDiskDetails.IdType
	deviceDetails.PartitionInfo.PartitionTableUUID = osDiskDetails.PartTableUUID

	deviceDetails.DependentDevices = osDiskDetails.Dependents

	diskInfo = append(diskInfo, deviceDetails)
	expectedEvent.eventDetails.Action = ""
	expectedEvent.eventDetails.Devices = diskInfo
	assert.Equal(t, expectedEvent, actualEvent)

	tests := map[string]struct {
		actualEvent   *event
		expectedEvent *event
	}{
		"match content of one event after process": {actualEvent: actualEvent, expectedEvent: expectedEvent},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedEvent, test.actualEvent)
		})
	}
}

func TestSend(t *testing.T) {
	event := newEvent()
	if event == nil {
		t.Error("event pointer should not be nil")
	}
	go event.send()
	msg := <-UdevEventMessageChannel

	tests := map[string]struct {
		actualEventMessage   controller.EventMessage
		expectedEventMessage controller.EventMessage
	}{
		"match content of one event after process": {actualEventMessage: msg, expectedEventMessage: event.eventDetails},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedEventMessage, test.actualEventMessage)
		})
	}
}
