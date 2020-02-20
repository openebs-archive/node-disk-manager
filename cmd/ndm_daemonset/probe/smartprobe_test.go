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

package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/smart"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockOsDiskToAPIBySmart() (apis.BlockDevice, error) {

	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		return apis.BlockDevice{}, err
	}

	mockOsDiskDetailsFromUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		return apis.BlockDevice{}, err
	}

	fakeDetails := apis.DeviceDetails{
		Compliance:       mockOsDiskDetails.Compliance,
		FirmwareRevision: mockOsDiskDetails.FirmwareRevision,
	}

	fakeCapacity := apis.DeviceCapacity{
		Storage:           mockOsDiskDetails.Capacity,
		LogicalSectorSize: mockOsDiskDetails.LBSize,
	}

	fakeObj := apis.DeviceSpec{
		Capacity:    fakeCapacity,
		Details:     fakeDetails,
		Path:        mockOsDiskDetails.DevPath,
		Partitioned: controller.NDMNotPartitioned,
	}

	devLinks := make([]apis.DeviceDevLink, 0)
	fakeObj.DevLinks = devLinks

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMBlockDeviceKind,
		APIVersion: controller.NDMVersion,
	}

	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockOsDiskDetailsFromUdev.Uid,
	}

	fakeDiskStatus := apis.DeviceStatus{
		State:      apis.BlockDeviceActive,
		ClaimState: apis.BlockDeviceUnclaimed,
	}

	fakeDr := apis.BlockDevice{
		TypeMeta:   fakeTypeMeta,
		ObjectMeta: fakeObjectMeta,
		Spec:       fakeObj,
		Status:     fakeDiskStatus,
	}

	return fakeDr, nil
}

func TestFillDiskDetailsBySmart(t *testing.T) {
	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	sProbe := smartProbe{}
	actualDiskInfo := &blockdevice.BlockDevice{}
	actualDiskInfo.DevPath = mockOsDiskDetails.DevPath
	sProbe.FillBlockDeviceDetails(actualDiskInfo)
	expectedDiskInfo := &blockdevice.BlockDevice{}
	expectedDiskInfo.DevPath = mockOsDiskDetails.DevPath
	expectedDiskInfo.Capacity.Storage = mockOsDiskDetails.Capacity
	expectedDiskInfo.Capacity.LogicalSectorSize = mockOsDiskDetails.LBSize
	expectedDiskInfo.DeviceDetails.FirmwareRevision = mockOsDiskDetails.FirmwareRevision
	expectedDiskInfo.DeviceDetails.Compliance = mockOsDiskDetails.Compliance
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}

func TestSmartProbe(t *testing.T) {
	mockOsDiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	mockOsDiskDetailsUsingUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}

	fakeHostName := "node-name"
	fakeNdmClient := CreateFakeClient(t)
	probes := make([]*controller.Probe, 0)
	filters := make([]*controller.Filter, 0)
	nodeAttributes := make(map[string]string)
	nodeAttributes[controller.HostNameKey] = fakeHostName
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		Clientset:      fakeNdmClient,
		Mutex:          mutex,
		Probes:         probes,
		Filters:        filters,
		NodeAttributes: nodeAttributes,
	}

	smartProbe := newSmartProbe("fakeController")
	var pi controller.ProbeInterface = smartProbe

	newRegisterProbe := &registerProbe{
		priority:   1,
		name:       "smart probe",
		state:      true,
		pi:         pi,
		controller: fakeController,
	}

	newRegisterProbe.register()

	// Add one filter
	filter := &alwaysTrueFilter{}
	filter1 := &controller.Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}

	fakeController.AddNewFilter(filter1)
	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}

	eventmsg := make([]*blockdevice.BlockDevice, 0)
	deviceDetails := &blockdevice.BlockDevice{}
	deviceDetails.UUID = mockOsDiskDetailsUsingUdev.Uid
	deviceDetails.DevPath = mockOsDiskDetails.DevPath
	eventmsg = append(eventmsg, deviceDetails)

	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addBlockDeviceEvent(eventDetails)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetBlockDevice(mockOsDiskDetailsUsingUdev.Uid)
	if err1 != nil {
		t.Fatal(err1)
	}

	fakeDr, err := mockOsDiskToAPIBySmart()
	if err != nil {
		t.Fatal(err)
	}
	fakeDr.ObjectMeta.Labels[controller.KubernetesHostNameLabel] = fakeController.NodeAttributes[controller.HostNameKey]
	fakeDr.ObjectMeta.Labels[controller.NDMDeviceTypeKey] = "blockdevice"
	fakeDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString

	tests := map[string]struct {
		actualDisk    apis.BlockDevice
		expectedDisk  apis.BlockDevice
		actualError   error
		expectedError error
	}{
		"add event for resource with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
