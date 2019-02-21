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
	"fmt"
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	mockuid        = "fake-disk-uid"
	ignoreDiskUuid = "ignore-disk-uuid"
	fakeHostName   = "node-name"
	fakeModel      = "fake-disk-model"
	fakeSerial     = "fake-disk-serial"
	fakeVendor     = "fake-disk-vendor"
	fakeDiskType   = "disk"
)

// mockEmptyDiskCr returns empty diskCr
func mockEmptyDiskCr() apis.Disk {
	fakeDr := apis.Disk{}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   mockuid,
	}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       controller.NDMDiskKind,
		APIVersion: controller.NDMVersion,
	}
	fakeDr.ObjectMeta = fakeObjectMeta
	fakeDr.TypeMeta = fakeTypeMeta
	fakeDr.Status.State = controller.NDMActive
	fakeDr.Spec.DevLinks = make([]apis.DiskDevLink, 0)
	return fakeDr
}

func CreateFakeClient(t *testing.T) client.Client {
	diskR := &apis.Disk{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummy-disk",
		},
	}

	diskList := &apis.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "",
		},
	}

	deviceR := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   "dummy-device",
		},
	}

	deviceList := &apis.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "",
		},
	}

	objs := []runtime.Object{diskR}
	s := scheme.Scheme
	s.AddKnownTypes(apis.SchemeGroupVersion, diskR)
	s.AddKnownTypes(apis.SchemeGroupVersion, diskList)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceR)
	s.AddKnownTypes(apis.SchemeGroupVersion, deviceList)

	fakeNdmClient := ndmFakeClientset.NewFakeClient(objs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	return fakeNdmClient
}

type fakeFilter struct{}

func (nf *fakeFilter) Start() {}

func (nf *fakeFilter) Include(fakeDiskInfo *controller.DiskInfo) bool {
	return true
}

func (nf *fakeFilter) Exclude(fakeDiskInfo *controller.DiskInfo) bool {
	return fakeDiskInfo.Uuid != ignoreDiskUuid
}

func TestAddEvent(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Mutex:         &sync.Mutex{},
		Filters:       make([]*controller.Filter, 0),
		Probes:        make([]*controller.Probe, 0),
	}
	//add one filter
	filter := &fakeFilter{}
	filter1 := &controller.Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}
	fakeController.AddNewFilter(filter1)
	// add one probe
	testProbe := &fakeProbe{}
	probe1 := &controller.Probe{
		Name:      "probe1",
		State:     true,
		Interface: testProbe,
	}
	fakeController.AddNewProbe(probe1)

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	// device-1 details
	eventmsg := make([]*controller.DiskInfo, 0)
	device1Details := &controller.DiskInfo{}
	device1Details.ProbeIdentifiers.Uuid = mockuid
	device1Details.DiskType = libudevwrapper.UDEV_DISK
	eventmsg = append(eventmsg, device1Details)
	// device-2 details
	device2Details := &controller.DiskInfo{}
	device2Details.ProbeIdentifiers.Uuid = ignoreDiskUuid
	device2Details.DiskType = libudevwrapper.UDEV_DISK
	eventmsg = append(eventmsg, device2Details)
	// Creating one event message
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: eventmsg,
	}
	probeEvent.addEvent(eventDetails)
	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(mockuid)

	// Retrieve disk resource
	cdr2, _ := fakeController.GetDisk(ignoreDiskUuid)
	if cdr2 != nil {
		t.Error("resource with ignoreDiskUuid should not be present in etcd")
	}
	// Create one fake disk resource
	fakeDr := mockEmptyDiskCr()
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[controller.NDMDiskTypeKey] = fakeDiskType
	fakeDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString
	fakeDr.Spec.Details.Model = fakeModel
	fakeDr.Spec.Details.Serial = fakeSerial
	fakeDr.Spec.Details.Vendor = fakeVendor

	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"resource with 'fake-disk-uid' uuid for create resource": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
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
	// Create one fake disk resource
	fakeDr := mockEmptyDiskCr()
	fakeDr.ObjectMeta.Labels[controller.NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[controller.NDMDiskTypeKey] = fakeDiskType
	fakeDr.ObjectMeta.Labels[controller.NDMManagedKey] = controller.TrueString
	fakeController.CreateDisk(fakeDr)

	probeEvent := &ProbeEvent{
		Controller: fakeController,
	}
	eventmsg := make([]*controller.DiskInfo, 0)
	deviceDetails := &controller.DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = mockuid
	deviceDetails.DiskType = libudevwrapper.UDEV_DISK
	eventmsg = append(eventmsg, deviceDetails)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_REMOVE,
		Devices: eventmsg,
	}
	probeEvent.deleteEvent(eventDetails)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(mockuid)

	fakeDr.Status.State = controller.NDMInactive
	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"remove resource with 'fake-disk-uid' uuid": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
