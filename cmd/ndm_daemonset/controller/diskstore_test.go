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
	"testing"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// mockEmptyDiskCr returns Disk object with minimum attributes it is used in unit test cases.
func mockEmptyDiskCr() apis.Disk {
	fakeDr := apis.Disk{}
	fakeObjectMeta := metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDiskUID,
	}
	fakeTypeMeta := metav1.TypeMeta{
		Kind:       NDMDiskKind,
		APIVersion: NDMVersion,
	}
	fakeDr.ObjectMeta = fakeObjectMeta
	fakeDr.TypeMeta = fakeTypeMeta
	fakeDr.Status.State = NDMActive
	fakeDr.Spec.DevLinks = make([]apis.DiskDevLink, 0)
	fakeDr.Spec.FileSystem.IsFormated = true
	return fakeDr
}

func TestCreateDisk(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create disk resource dr1
	dr1 := fakeDr
	dr1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr1.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr1.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr1)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Create resource which is already present, it should update
	dr2 := fakeDr
	dr2.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr2.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr2.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr2)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(fakeDiskUID)

	// Create disk resource dr3
	dr3 := newFakeDr
	dr3.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr3.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr3.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr3)

	// Retrieve disk resource
	cdr3, err3 := fakeController.GetDisk(newFakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"create one resource":                          {actualDisk: *cdr1, actualError: err1, expectedDisk: dr1, expectedError: nil},
		"create one resource which is already present": {actualDisk: *cdr2, actualError: err2, expectedDisk: dr2, expectedError: nil},
		"create another resource":                      {actualDisk: *cdr3, actualError: err3, expectedDisk: dr3, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestUpdateDisk(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Update a resource which is not present
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	err := fakeController.UpdateDisk(dr, nil)
	if err == nil {
		t.Error("error should not be nil as the resource is not present")
	}

	// Create one disk resource and update it.
	fakeController.CreateDisk(dr)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Update already created resource
	err = fakeController.UpdateDisk(dr, cdr1)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(fakeDiskUID)

	// Pass nil value in old resource
	err = fakeController.UpdateDisk(dr, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve disk resource
	cdr3, err3 := fakeController.GetDisk(fakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"create one resource":                        {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"update resource when old value present":     {actualDisk: *cdr2, actualError: err2, expectedDisk: dr, expectedError: nil},
		"update resource when old value not present": {actualDisk: *cdr3, actualError: err3, expectedDisk: dr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}

	// Retrieve disk resource
	cdr, err := fakeController.GetDisk(fakeDiskUID)

	dr.ObjectMeta.Name = "disk-updated-fake-uuid"
	err = fakeController.UpdateDisk(dr, cdr)
	if err == nil {
		t.Error("if resource is not present then it should return error")
	}
}

func TestDeactivateDisk(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create one resource and deactivate it.
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)
	fakeController.DeactivateDisk(dr)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Deactivate one resource which is not present it should return error
	dr1 := newFakeDr
	dr1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr1.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr1.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.DeactivateDisk(dr1)

	// Create another resource and deactivate it.
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	newDr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(newDr)
	fakeController.DeactivateDisk(newDr)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(newFakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"deactivate dr resource":    {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"deactivate newDr resource": {actualDisk: *cdr2, actualError: err2, expectedDisk: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.expectedDisk.Status.State = NDMInactive
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeleteDisk(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create one resource and delete it
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)
	fakeController.DeleteDisk(fakeDiskUID)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Delete one resource which is not present it should return error
	fakeController.DeleteDisk("another-uuid")

	// Create another resource and delete it
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	newDr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(newDr)
	fakeController.DeleteDisk(newFakeDiskUID)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(fakeDiskUID)

	tests := map[string]struct {
		expectedError error
		actualDisk    *apis.Disk
		expectedDisk  *apis.Disk
	}{
		"delete dr resource":    {expectedError: err1, actualDisk: cdr1, expectedDisk: nil},
		"delete newDr resource": {expectedError: err2, actualDisk: cdr2, expectedDisk: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.expectedError == nil {
				t.Error("error should not be nil as the resource was deleted")
			}
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}

func TestListDiskResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create disk resource dr
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)

	// We have created this device as part of fake.client creation
	// so delete it here
	fakeController.DeleteDisk("dummy-disk")

	// Create disk resource newDr
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	newDr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(newDr)
	listDevice, err := fakeController.ListDiskResource()

	// TypeMeta should be same
	typeMeta := newFakeTypeMeta
	listMeta := metav1.ListMeta{}
	diskList := make([]apis.Disk, 0)
	diskList = append(diskList, dr)
	diskList = append(diskList, newDr)
	expectedList := &apis.DiskList{
		TypeMeta: typeMeta,
		ListMeta: listMeta,
		Items:    diskList,
	}

	tests := map[string]struct {
		actualError      error
		expectedError    error
		actualDiskList   *apis.DiskList
		expectedDiskList *apis.DiskList
	}{
		"delete dr resource": {actualError: err, expectedError: nil, actualDiskList: listDevice, expectedDiskList: expectedList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDiskList, test.actualDiskList)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestGetExistingDiskResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create disk resource dr
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)

	// Create disk resource newDr
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	newDr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(newDr)

	listDr, err := fakeController.ListDiskResource()
	if err != nil {
		t.Fatal(err)
	}
	cdr1 := fakeController.GetExistingDiskResource(listDr, fakeDiskUID)
	cdr2 := fakeController.GetExistingDiskResource(listDr, newFakeDiskUID)
	cdr3 := fakeController.GetExistingDiskResource(listDr, "newFakeDiskUID")
	tests := map[string]struct {
		actualDisk   *apis.Disk
		expectedDisk *apis.Disk
	}{
		"resource with 'fake-disk-uid' uuid":             {actualDisk: cdr1, expectedDisk: &dr},
		"resource with 'new-fake-disk-uid' uuid":         {actualDisk: cdr2, expectedDisk: &newDr},
		"resource with invalid uuid not present in etcd": {actualDisk: cdr3, expectedDisk: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}

/*
 * PushResource take 2 argument one is old disk resource and other is
 * DiskInfo struct if old disk resource is not present it creates one
 * new resource if not then it update that resource with updated DiskInfo
 */
func TestPushDiskResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// create one DiskInfo struct with mock uuid
	deviceDetails := &DiskInfo{}
	deviceDetails.ProbeIdentifiers.Uuid = fakeDiskUID
	deviceDetails.DiskType = NDMDefaultDiskType

	// Create one fake Disk struct
	fakeDr := mockEmptyDiskCr()
	fakeDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	fakeDr.ObjectMeta.Labels[NDMManagedKey] = TrueString

	// Pass 1st argument as nil then it creates one disk resource
	fakeController.PushDiskResource(nil, deviceDetails)

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Pass old disk resource as 1st argument then it updates resource
	fakeController.PushDiskResource(cdr1, deviceDetails)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(fakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		expectedDisk  apis.Disk
		actualError   error
		expectedError error
	}{
		"push resource with 'fake-disk-uid' uuid for create resource": {actualDisk: *cdr1, expectedDisk: fakeDr, actualError: err1, expectedError: nil},
		"push resource with 'fake-disk-uid' uuid for update resource": {actualDisk: *cdr2, expectedDisk: fakeDr, actualError: err2, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeactivateStaleDiskResource(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// Create disk resource dr
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)

	// Create disk resource newDr
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	newDr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	newDr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(newDr)

	// Add one resource's uuid so state of the other resource should be inactive.
	diskList := make([]string, 0)
	diskList = append(diskList, newFakeDiskUID)
	fakeController.DeactivateStaleDiskResource(diskList)
	dr.Status.State = NDMInactive

	// Retrieve disk resource
	cdr1, err1 := fakeController.GetDisk(fakeDiskUID)

	// Retrieve disk resource
	cdr2, err2 := fakeController.GetDisk(newFakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"resource1 present in etcd but not in system": {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"resource2 present in both etcd and system":   {actualDisk: *cdr2, actualError: err2, expectedDisk: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestMarkDiskStatusToUnknown(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	dr := mockEmptyDiskCr()
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	dr.ObjectMeta.Labels[NDMDiskTypeKey] = NDMDefaultDiskType
	dr.ObjectMeta.Labels[NDMManagedKey] = TrueString
	fakeController.CreateDisk(dr)

	fakeController.MarkDiskStatusToUnknown()
	dr.Status.State = NDMUnknown

	// Retrieve disk resource
	cdr, err := fakeController.GetDisk(fakeDiskUID)

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"DeactivateOwnedDiskResource should make all present resources state inactive": {actualDisk: *cdr, actualError: err, expectedDisk: dr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
