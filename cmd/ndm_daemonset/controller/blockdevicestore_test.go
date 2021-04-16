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

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockEmptyDeviceCr returns BlockDevice object with minimum attributes it is used in unit test cases.
func mockEmptyDeviceCr() apis.BlockDevice {
	fakeDevice := apis.BlockDevice{}

	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDeviceUID,
	}

	fakeTypeMeta := metav1.TypeMeta{
		Kind:       NDMBlockDeviceKind,
		APIVersion: NDMVersion,
	}

	fakeDeviceSpec := apis.DeviceSpec{
		DevLinks:    make([]apis.DeviceDevLink, 0),
		Partitioned: NDMNotPartitioned,
	}

	fakeDevice.ObjectMeta = fakeObjectMeta
	fakeDevice.TypeMeta = fakeTypeMeta
	fakeDevice.Status.ClaimState = apis.BlockDeviceUnclaimed
	fakeDevice.Status.State = NDMActive
	fakeDevice.Spec = fakeDeviceSpec
	/*.Partitioned = NDMNotPartitioned
	fakeDevice.Spec.DevLinks = make([]apis.DeviceDevLink, 0)*/
	return fakeDevice
}

func TestCreateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create blockdevice resource devR1
	devR1 := fakeDevice
	devR1.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	devR1.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(devR1)

	// Retrieve blockdevice resource
	cdevR1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Create resource which is already present, it should update
	devR2 := fakeDevice
	devR2.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	devR2.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(devR2)

	// Retrieve blockdevice resource
	cdevR2, err2 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Create blockdevice resource devR3
	devR3 := newFakeDevice
	devR3.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	devR3.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(devR3)

	// Retrieve blockdevice resource
	cdevR3, err3 := fakeController.GetBlockDevice(newFakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		actualError    error
		expectedDevice apis.BlockDevice
		expectedError  error
	}{
		"create one resource":                          {actualDevice: *cdevR1, actualError: err1, expectedDevice: devR1, expectedError: nil},
		"create one resource which is already present": {actualDevice: *cdevR2, actualError: err2, expectedDevice: devR2, expectedError: nil},
		"create another resource":                      {actualDevice: *cdevR3, actualError: err3, expectedDevice: devR3, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Update a resource which is not present
	devR := fakeDevice
	devR.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	devR.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	err := fakeController.UpdateBlockDevice(devR, nil)
	if err == nil {
		t.Error("error should not be nil as the resource is not present")
	}

	// Create one blockdevice resource and update it.
	fakeController.CreateBlockDevice(devR)

	// Retrieve blockdevice resource
	cdevR1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Update already created resource
	err = fakeController.UpdateBlockDevice(devR, cdevR1)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve blockdevice resource
	cdevR2, err2 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Pass nil value in old resource
	err = fakeController.UpdateBlockDevice(devR, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve blockdevice resource
	cdevR3, err3 := fakeController.GetBlockDevice(fakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		actualError    error
		expectedDevice apis.BlockDevice
		expectedError  error
	}{
		"create one resource":                          {actualDevice: *cdevR1, actualError: err1, expectedDevice: devR, expectedError: nil},
		"create one resource which is already present": {actualDevice: *cdevR2, actualError: err2, expectedDevice: devR, expectedError: nil},
		"create another resource":                      {actualDevice: *cdevR3, actualError: err3, expectedDevice: devR, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}

	// Retrieve disk resource
	cdevR, err := fakeController.GetBlockDevice(fakeDeviceUID)
	if err != nil {
		t.Fatal(err)
	}
	cdevR.ObjectMeta.Name = "disk-updated-fake-uuid"
	err = fakeController.UpdateBlockDevice(*cdevR, nil)
	if err == nil {
		t.Error("if resource is not present then it should return error")
	}
}

func TestDeactivateDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create one resource and deactivate it.
	dr := fakeDevice
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)
	fakeController.DeactivateBlockDevice(dr)

	// Retrieve blockdevice resource
	cdr1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Deactivate one resource which is not present it should return error
	dr1 := newFakeDevice
	dr1.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr1.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.DeactivateBlockDevice(dr1)

	// Create another resource and deactivate it.
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(newDr)
	fakeController.DeactivateBlockDevice(newDr)

	// Retrieve blockdevice resource
	cdr2, err2 := fakeController.GetBlockDevice(newFakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		actualError    error
		expectedDevice apis.BlockDevice
		expectedError  error
	}{
		"deactivate dr resource":    {actualDevice: *cdr1, actualError: err1, expectedDevice: dr, expectedError: nil},
		"deactivate newDr resource": {actualDevice: *cdr2, actualError: err2, expectedDevice: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.expectedDevice.Status.State = NDMInactive
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create one resource and delete it
	dr := fakeDevice
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)
	fakeController.DeleteBlockDevice(fakeDeviceUID)

	// Retrieve blockdevice resource
	cdr1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Delete one resource which is not present it should return error
	fakeController.DeleteBlockDevice("another-uuid")

	// Create another resource and delete it
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(newDr)
	fakeController.DeleteBlockDevice(newFakeDeviceUID)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetBlockDevice(fakeDeviceUID)

	tests := map[string]struct {
		expectedError  error
		actualDevice   *apis.BlockDevice
		expectedDevice *apis.BlockDevice
	}{
		"delete dr resource":    {expectedError: err1, actualDevice: cdr1, expectedDevice: nil},
		"delete newDr resource": {expectedError: err2, actualDevice: cdr2, expectedDevice: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.expectedError == nil {
				t.Error("error should not be nil as the resource was deleted")
			}
			if test.expectedDevice == nil {
				assert.Equal(t, test.expectedDevice, test.actualDevice)
			} else {
				// compare only if not nil
				compareBlockDevice(t, *test.expectedDevice, *test.actualDevice)
			}
		})
	}
}

func TestListDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create blockdevice resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)

	// Create blockdevice resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(newDr)
	listDevice, err := fakeController.ListBlockDeviceResource(false)

	// TypeMeta should be same
	typeMeta := newFakeDeviceTypeMeta
	listMeta := metav1.ListMeta{}
	deviceList := make([]apis.BlockDevice, 0)
	deviceList = append(deviceList, dr)
	deviceList = append(deviceList, newDr)
	expectedList := &apis.BlockDeviceList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    deviceList,
	}

	tests := map[string]struct {
		actualError        error
		expectedError      error
		actualDeviceList   *apis.BlockDeviceList
		expectedDeviceList *apis.BlockDeviceList
	}{
		"delete dr resource": {actualError: err, expectedError: nil, actualDeviceList: listDevice, expectedDeviceList: expectedList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			compareBlockDeviceList(t, *test.expectedDeviceList, *test.actualDeviceList)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestGetExistingDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create blockdevice resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)

	// Create blockdevice resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(newDr)

	listDr, err := fakeController.ListBlockDeviceResource(false)
	if err != nil {
		t.Fatal(err)
	}
	cdr1 := fakeController.GetExistingBlockDeviceResource(listDr, fakeDeviceUID)
	cdr2 := fakeController.GetExistingBlockDeviceResource(listDr, newFakeDeviceUID)
	cdr3 := fakeController.GetExistingBlockDeviceResource(listDr, "newFakeDeviceUID")
	tests := map[string]struct {
		actualDevice   *apis.BlockDevice
		expectedDevice *apis.BlockDevice
	}{
		"resource with 'fake-blockdevice-uid' uuid":      {actualDevice: cdr1, expectedDevice: &dr},
		"resource with 'new-fake-blockdevice-uid' uuid":  {actualDevice: cdr2, expectedDevice: &newDr},
		"resource with invalid uuid not present in etcd": {actualDevice: cdr3, expectedDevice: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.expectedDevice == nil {
				assert.Equal(t, test.expectedDevice, test.actualDevice)
			} else {
				// compare only if not nil
				compareBlockDevice(t, *test.expectedDevice, *test.actualDevice)
			}
		})
	}
}

/*
 * PushBlockDeviceResource take 2 argument one is old blockdevice resource and other is
 * DeviceInfo struct. If old blockdevice resource is not present it creates one
 * new resource, if not then it update that resource with updated DeviceInfo
 */
func TestPushDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create one DeviceInfo struct with mock uuid
	deviceDetails := &DeviceInfo{}
	deviceDetails.NodeAttributes = make(map[string]string)
	deviceDetails.UUID = fakeDeviceUID

	// Create one fake BlockDevice struct
	fakeDr := mockEmptyDeviceCr()
	fakeDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	fakeDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeDr.ObjectMeta.Labels[NDMManagedKey] = TrueString

	// Pass 1st argument as nil then it creates one disk resource
	fakeController.PushBlockDeviceResource(nil, deviceDetails)

	// Retrieve blockdevice resource
	cdr1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Pass old blockdevice resource as 1st argument then it updates resource
	fakeController.PushBlockDeviceResource(cdr1, deviceDetails)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetBlockDevice(fakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		expectedDevice apis.BlockDevice
		actualError    error
		expectedError  error
	}{
		"push resource with 'fake-blockdevice-uid' uuid for create resource": {actualDevice: *cdr1, expectedDevice: fakeDr, actualError: err1, expectedError: nil},
		"push resource with 'fake-blockdevice-uid' uuid for update resource": {actualDevice: *cdr2, expectedDevice: fakeDr, actualError: err2, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeactivateStaleDeviceResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	// Create blockdevice resource dr
	dr := fakeDevice
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)

	// Create blockdevice resource newDr
	newDr := newFakeDevice
	newDr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	newDr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(newDr)

	// Add one resource's uuid so state of the other resource should be inactive.
	deviceList := make([]string, 0)
	deviceList = append(deviceList, newFakeDeviceUID)
	fakeController.DeactivateStaleBlockDeviceResource(deviceList)
	dr.Status.State = NDMInactive

	// Retrieve blockdevice resource
	cdr1, err1 := fakeController.GetBlockDevice(fakeDeviceUID)

	// Retrieve blockdevice resource
	cdr2, err2 := fakeController.GetBlockDevice(newFakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		actualError    error
		expectedDevice apis.BlockDevice
		expectedError  error
	}{
		"resource1 present in etcd but not in system": {actualDevice: *cdr1, actualError: err1, expectedDevice: dr, expectedError: nil},
		"resource2 present in both etcd and system":   {actualDevice: *cdr2, actualError: err2, expectedDevice: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestMarkDeviceStatusToUnknown(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[HostNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}

	dr := mockEmptyDeviceCr()
	dr.ObjectMeta.Labels[KubernetesHostNameLabel] = fakeController.NodeAttributes[HostNameKey]
	dr.ObjectMeta.Labels[NDMDeviceTypeKey] = NDMDefaultDeviceType
	fakeController.CreateBlockDevice(dr)

	fakeController.MarkBlockDeviceStatusToUnknown()
	dr.Status.State = NDMUnknown

	// Retrieve blockdevice resource
	cdr, err := fakeController.GetBlockDevice(fakeDeviceUID)

	tests := map[string]struct {
		actualDevice   apis.BlockDevice
		actualError    error
		expectedDevice apis.BlockDevice
		expectedError  error
	}{
		"DeactivateOwnedDeviceResource should make all present resources state inactive": {actualDevice: *cdr, actualError: err, expectedDevice: dr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			compareBlockDevice(t, test.expectedDevice, test.actualDevice)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

// compareBlockDevice is the custom blockdevice comparison function. Only those values that need to be checked
// for equality will be checked here. Resource version field will not be checked as it
// will be updated on every write. Refer https://github.com/kubernetes-sigs/controller-runtime/pull/620
func compareBlockDevice(t *testing.T, bd1, bd2 apis.BlockDevice) {
	assert.Equal(t, bd1.Name, bd2.Name)
	assert.Equal(t, bd1.Labels, bd2.Labels)
	assert.Equal(t, bd1.Spec, bd2.Spec)
	assert.Equal(t, bd1.Status, bd2.Status)
}

// compareBlockDeviceList is the custom comparison function for blockdevice list
func compareBlockDeviceList(t *testing.T, bdList1, bdList2 apis.BlockDeviceList) {
	assert.Equal(t, len(bdList1.Items), len(bdList2.Items))
	for i := 0; i < len(bdList2.Items); i++ {
		compareBlockDevice(t, bdList1.Items[i], bdList2.Items[i])
	}
}
