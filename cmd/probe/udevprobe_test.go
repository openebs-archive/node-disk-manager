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
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type alwaysTrueFilter struct{}

func (nf *alwaysTrueFilter) Start() {}

func (nf *alwaysTrueFilter) Include(fakeDiskInfo *controller.DiskInfo) bool {
	return true
}

func (nf *alwaysTrueFilter) Exclude(fakeDiskInfo *controller.DiskInfo) bool {
	return true
}

func mockOsDiskToAPI() (apis.Disk, error) {
	mockOsDiskDetails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		return apis.Disk{}, err
	}
	fakeDetails := apis.DiskDetails{
		Model:  mockOsDiskDetails.Model,
		Serial: mockOsDiskDetails.Serial,
		Vendor: mockOsDiskDetails.Vendor,
	}
	fakeObj := apis.DiskSpec{
		Path:    mockOsDiskDetails.DevNode,
		Details: fakeDetails,
	}

	devLinks := make([]apis.DiskDevLink, 0)
	if len(mockOsDiskDetails.ByIdDevLinks) != 0 {
		byIdLinks := apis.DiskDevLink{
			Kind:  "by-id",
			Links: mockOsDiskDetails.ByIdDevLinks,
		}
		devLinks = append(devLinks, byIdLinks)
	}
	if len(mockOsDiskDetails.ByPathDevLinks) != 0 {
		byPathLinks := apis.DiskDevLink{
			Kind:  "by-path",
			Links: mockOsDiskDetails.ByPathDevLinks,
		}
		devLinks = append(devLinks, byPathLinks)
	}
	fakeObj.DevLinks = devLinks

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMKind,
		APIVersion: controller.NDMVersion,
	}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockOsDiskDetails.Uid,
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

func TestFillDiskDetails(t *testing.T) {
	mockOsDiskDetails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	uProbe := udevProbe{}
	actualDiskInfo := controller.NewDiskInfo()
	actualDiskInfo.ProbeIdentifiers.UdevIdentifier = mockOsDiskDetails.SysPath
	uProbe.FillDiskDetails(actualDiskInfo)
	expectedDiskInfo := &controller.DiskInfo{}
	expectedDiskInfo.ProbeIdentifiers.UdevIdentifier = mockOsDiskDetails.SysPath
	expectedDiskInfo.ProbeIdentifiers.SmartIdentifier = mockOsDiskDetails.DevNode
	expectedDiskInfo.ProbeIdentifiers.SeachestIdentifier = mockOsDiskDetails.DevNode
	expectedDiskInfo.Model = mockOsDiskDetails.Model
	expectedDiskInfo.Path = mockOsDiskDetails.DevNode
	expectedDiskInfo.Serial = mockOsDiskDetails.Serial
	expectedDiskInfo.Vendor = mockOsDiskDetails.Vendor
	expectedDiskInfo.DiskType = "disk"
	expectedDiskInfo.ByIdDevLinks = mockOsDiskDetails.ByIdDevLinks
	expectedDiskInfo.ByPathDevLinks = mockOsDiskDetails.ByPathDevLinks
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}

func TestUdevProbe(t *testing.T) {
	mockOsDiskDetails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	fakeHostName := "node-name"
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	probes := make([]*controller.Probe, 0)
	filters := make([]*controller.Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Mutex:         mutex,
		Probes:        probes,
		Filters:       filters,
	}
	udevprobe := newUdevProbe(fakeController)
	var pi controller.ProbeInterface = udevprobe
	newRegisterProbe := &registerProbe{
		priority:   1,
		name:       "udev probe",
		state:      true,
		pi:         pi,
		controller: fakeController,
	}
	newRegisterProbe.register()
	//add one filter
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
	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockOsDiskDetails.Uid
	deviceDetails.ProbeIdentifiers.UdevIdentifier = mockOsDiskDetails.SysPath
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addDiskEvent(eventDetails)
	fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockOsDiskDetails.Uid, metav1.GetOptions{})
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(mockOsDiskDetails.Uid, metav1.GetOptions{})
	fakeDr, err := mockOsDiskToAPI()
	if err != nil {
		t.Fatal(err)
	}
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[controller.NDMDiskTypeKey] = "disk"
	fakeDr.ObjectMeta.Labels[controller.NDMUnmanagedKey] = controller.FalseString
	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
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
