/*
Copyright 2019 OpenEBS Authors.

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
	"testing"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// mockEmptyDeviceCr returns Device object with minimum attributes it is used in unit test cases.
func mockEmptyDeviceCr() apis.Device {
	fakeDevice := apis.Device{}

	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDeviceUid,
	}

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       NDMDeviceKind,
		APIVersion: NDMVersion,
	}

	fakeDeviceSpec := apis.DeviceSpec{
		DevLinks:    make([]apis.DeviceDevLink, 0),
		Partitioned: NDMNotPartitioned,
	}

	fakeDevice.ObjectMeta = fakeObjectMeta
	fakeDevice.TypeMeta = fakeTypeMeta
	fakeDevice.ClaimState.State = NDMUnclaimed
	fakeDevice.Status.State = NDMActive
	fakeDevice.Spec = fakeDeviceSpec
	/*.Partitioned = NDMNotPartitioned
	fakeDevice.Spec.DevLinks = make([]apis.DeviceDevLink, 0)*/
	return fakeDevice
}

func TestCreateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create device resource devR1
	devR1 := fakeDevice
	devR1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	devR1.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(devR1)

	// Retrieve device resource
	cdevR1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Create resource which is already present, it should update
	devR2 := fakeDevice
	devR2.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	devR2.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(devR2)

	// Retrieve device resource
	cdevR2, err2 := fakeController.GetDevice(fakeDeviceUid)

	// Create device resource devR3
	devR3 := newFakeDevice
	devR3.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	devR3.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(devR3)

	// Retrieve device resource
	cdevR3, err3 := fakeController.GetDevice(newFakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		actualError    error
		expectedDevice apis.Device
		expectedError  error
	}{
		"create one resource":                          {actualDevice: *cdevR1, actualError: err1, expectedDevice: devR1, expectedError: nil},
		"create one resource which is already present": {actualDevice: *cdevR2, actualError: err2, expectedDevice: devR2, expectedError: nil},
		"create another resource":                      {actualDevice: *cdevR3, actualError: err3, expectedDevice: devR3, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Update a resource which is not present
	devR := fakeDevice
	devR.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	devR.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	err := fakeController.UpdateDevice(devR, nil)
	if err == nil {
		t.Error("error should not be nil as the resource is not present")
	}

	// Create one device resource and update it.
	fakeController.CreateDevice(devR)

	// Retrieve device resource
	cdevR1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Update already created resource
	err = fakeController.UpdateDevice(devR, cdevR1)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve device resource
	cdevR2, err2 := fakeController.GetDevice(fakeDeviceUid)

	// Pass nil value in old resource
	err = fakeController.UpdateDevice(devR, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve device resource
	cdevR3, err3 := fakeController.GetDevice(fakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		actualError    error
		expectedDevice apis.Device
		expectedError  error
	}{
		"create one resource":                          {actualDevice: *cdevR1, actualError: err1, expectedDevice: devR, expectedError: nil},
		"create one resource which is already present": {actualDevice: *cdevR2, actualError: err2, expectedDevice: devR, expectedError: nil},
		"create another resource":                      {actualDevice: *cdevR3, actualError: err3, expectedDevice: devR, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}

	// Retrieve disk resource
	cdevR, err := fakeController.GetDevice(fakeDeviceUid)

	devR.ObjectMeta.Name = "disk-updated-fake-uuid"
	err = fakeController.UpdateDevice(devR, cdevR)
	if err == nil {
		t.Error("if resource is not present then it should return error")
	}
}

func TestDeactivateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create one resource and deactivate it.
	dr := fakeDevice
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)
	fakeController.DeactivateDevice(dr)

	// Retrieve device resource
	cdr1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Deactivate one resource which is not present it should return error
	dr1 := newFakeDevice
	dr1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr1.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.DeactivateDevice(dr1)

	// Create another resource and deactivate it.
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(newDr)
	fakeController.DeactivateDevice(newDr)

	// Retrieve device resource
	cdr2, err2 := fakeController.GetDevice(newFakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		actualError    error
		expectedDevice apis.Device
		expectedError  error
	}{
		"deactivate dr resource":    {actualDevice: *cdr1, actualError: err1, expectedDevice: dr, expectedError: nil},
		"deactivate newDr resource": {actualDevice: *cdr2, actualError: err2, expectedDevice: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.expectedDevice.Status.State = NDMInactive
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create one resource and delete it
	dr := fakeDevice
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)
	fakeController.DeleteDevice(fakeDeviceUid)

	// Retrieve device resource
	cdr1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Delete one resource which is not present it should return error
	fakeController.DeleteDevice("another-uuid")

	// Create another resource and delete it
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(newDr)
	fakeController.DeleteDevice(newFakeDeviceUid)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDevice(fakeDeviceUid)

	tests := map[string]struct {
		expectedError  error
		actualDevice   *apis.Device
		expectedDevice *apis.Device
	}{
		"delete dr resource":    {expectedError: err1, actualDevice: cdr1, expectedDevice: nil},
		"delete newDr resource": {expectedError: err2, actualDevice: cdr2, expectedDevice: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.expectedError == nil {
				t.Error("error should not be nil as the resource was deleted")
			}
			assert.Equal(t, test.expectedDevice, test.actualDevice)
		})
	}
}

func TestListDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create device resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)

	// Create device resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(newDr)
	listDevice, err := fakeController.ListDeviceResource()

	// TypeMeta should be same
	typeMeta := newFakeDeviceTypeMeta
	listMeta := metav1.ListMeta{}
	deviceList := make([]apis.Device, 0)
	deviceList = append(deviceList, dr)
	deviceList = append(deviceList, newDr)
	expectedList := &apis.DeviceList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    deviceList,
	}

	tests := map[string]struct {
		actualError        error
		expectedError      error
		actualDeviceList   *apis.DeviceList
		expectedDeviceList *apis.DeviceList
	}{
		"delete dr resource": {actualError: err, expectedError: nil, actualDeviceList: listDevice, expectedDeviceList: expectedList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDeviceList, test.actualDeviceList)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestGetExistingDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create device resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)

	// Create device resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(newDr)

	listDr, err := fakeController.ListDeviceResource()
	if err != nil {
		t.Fatal(err)
	}
	cdr1 := fakeController.GetExistingDeviceResource(listDr, fakeDeviceUid)
	cdr2 := fakeController.GetExistingDeviceResource(listDr, newFakeDeviceUid)
	cdr3 := fakeController.GetExistingDeviceResource(listDr, "newFakeDeviceUid")
	tests := map[string]struct {
		actualDevice   *apis.Device
		expectedDevice *apis.Device
	}{
		"resource with 'fake-device-uid' uuid":           {actualDevice: cdr1, expectedDevice: &dr},
		"resource with 'new-fake-device-uid' uuid":       {actualDevice: cdr2, expectedDevice: &newDr},
		"resource with invalid uuid not present in etcd": {actualDevice: cdr3, expectedDevice: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
		})
	}
}

/*
 * PushDeviceResource take 2 argument one is old device resource and other is
 * DeviceInfo struct. If old device resource is not present it creates one
 * new resource, if not then it update that resource with updated DeviceInfo
 */
func TestPushDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create one DeviceInfo struct with mock uuid
	deviceDetails := &DeviceInfo{}
	deviceDetails.Uuid = fakeDeviceUid

	// Create one fake Device struct
	fakeDr := mockEmptyDeviceCr()
	fakeDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType

	// Pass 1st argument as nil then it creates one disk resource
	fakeController.PushDeviceResource(nil, deviceDetails)

	// Retrieve device resource
	cdr1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Pass old device resource as 1st argument then it updates resource
	fakeController.PushDeviceResource(cdr1, deviceDetails)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDevice(fakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		expectedDevice apis.Device
		actualError    error
		expectedError  error
	}{
		"push resource with 'fake-device-uid' uuid for create resource": {actualDevice: *cdr1, expectedDevice: fakeDr, actualError: err1, expectedError: nil},
		"push resource with 'fake-device-uid' uuid for update resource": {actualDevice: *cdr2, expectedDevice: fakeDr, actualError: err2, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeactivateStaleDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create device resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)

	// Create device resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(newDr)

	// Add one resource's uuid so state of the other resource should be inactive.
	deviceList := make([]string, 0)
	deviceList = append(deviceList, newFakeDeviceUid)
	fakeController.DeactivateStaleDeviceResource(deviceList)
	dr.Status.State = NDMInactive

	// Retrieve device resource
	cdr1, err1 := fakeController.GetDevice(fakeDeviceUid)

	// Retrieve device resource
	cdr2, err2 := fakeController.GetDevice(newFakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		actualError    error
		expectedDevice apis.Device
		expectedError  error
	}{
		"resource1 present in etcd but not in system": {actualDevice: *cdr1, actualError: err1, expectedDevice: dr, expectedError: nil},
		"resource2 present in both etcd and system":   {actualDevice: *cdr2, actualError: err2, expectedDevice: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestMarkDeviceStatusToUnknown(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	dr := mockEmptyDeviceCr()
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateDevice(dr)

	fakeController.MarkDeviceStatusToUnknown()
	dr.Status.State = NDMUnknown

	// Retrieve device resource
	cdr, err := fakeController.GetDevice(fakeDeviceUid)

	tests := map[string]struct {
		actualDevice   apis.Device
		actualError    error
		expectedDevice apis.Device
		expectedError  error
	}{
		"DeactivateOwnedDeviceResource should make all present resources state inactive": {actualDevice: *cdr, actualError: err, expectedDevice: dr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
