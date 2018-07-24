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
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs.io/v1alpha1"
	ndmFakeClientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned/fake"
	smart "github.com/openebs/node-disk-manager/pkg/smart"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mockOsDiskToAPIBySmart() (apis.Disk, error) {

	mockOsdiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		return apis.Disk{}, err
	}

	mockOsdiskDetailsFromUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		return apis.Disk{}, err
	}

	fakeDetails := apis.DiskDetails{
		SPCVersion:       mockOsdiskDetails.SPCVersion,
		FirmwareRevision: mockOsdiskDetails.FirmwareRevision,
	}

	fakeCapacity := apis.DiskCapacity{
		Storage:           mockOsdiskDetails.Capacity,
		LogicalSectorSize: mockOsdiskDetails.LBSize,
	}

	fakeObj := apis.DiskSpec{
		Capacity: fakeCapacity,
		Details:  fakeDetails,
	}

	devLinks := make([]apis.DiskDevLink, 0)
	fakeObj.DevLinks = devLinks

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMKind,
		APIVersion: controller.NDMVersion,
	}

	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockOsdiskDetailsFromUdev.Uid,
	}

	fakeDiskStatus := apis.DiskStatus{
		State: controller.NDMActive,
	}

	fakeDr := apis.Disk{
		TypeMeta:   fakeTypeMeta,
		ObjectMeta: fakeObjectMeta,
		Spec:       fakeObj,
		Status:     fakeDiskStatus,
	}

	return fakeDr, nil
}

func TestFillDiskDetailsBySmart(t *testing.T) {
	mockOsdiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	sProbe := smartProbe{}
	actualDiskInfo := controller.NewDiskInfo()
	actualDiskInfo.ProbeIdentifiers.SmartIdentifier = mockOsdiskDetails.DevPath
	sProbe.FillDiskDetails(actualDiskInfo)
	expectedDiskInfo := &controller.DiskInfo{}
	expectedDiskInfo.ProbeIdentifiers.SmartIdentifier = mockOsdiskDetails.DevPath
	expectedDiskInfo.Capacity = mockOsdiskDetails.Capacity
	expectedDiskInfo.FirmwareRevision = mockOsdiskDetails.FirmwareRevision
	expectedDiskInfo.SPCVersion = mockOsdiskDetails.SPCVersion
	expectedDiskInfo.LogicalSectorSize = mockOsdiskDetails.LBSize
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}

func TestSmartProbe(t *testing.T) {
	mockOsdiskDetails, err := smart.MockScsiBasicDiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	mockOsdiskDetailsUsingUdev, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}

	fakeHostName := "node-name"
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	probes := make([]*controller.Probe, 0)
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Probes:        probes,
		Mutex:         mutex,
	}

	smartProbe := newSmartProbe("fakeController")
	var pi controller.ProbeInterface = smartProbe

	newPrgisterProbe := &registerProbe{
		priority:       1,
		probeName:      "smart probe",
		probeState:     true,
		probeInterface: pi,
		controller:     fakeController,
	}
	newPrgisterProbe.register()

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}

	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockOsdiskDetailsUsingUdev.Uid
	deviceDetails.ProbeIdentifiers.SmartIdentifier = mockOsdiskDetails.DevPath
	eventmsg = append(eventmsg, deviceDetails)

	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addDiskEvent(eventDetails)

	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockOsdiskDetailsUsingUdev.Uid, metav1.GetOptions{})
	if err1 != nil {
		t.Fatal(err1)
	}

	fakeDr, err := mockOsDiskToAPIBySmart()
	if err != nil {
		t.Fatal(err)
	}
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName

	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"add event for resouce with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
