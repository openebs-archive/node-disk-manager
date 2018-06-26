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

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs.io/v1alpha1"
	ndmFakeClientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateDisk(t *testing.T) {
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}

	// create resource1
	dr1 := fakeDr
	dr1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr1)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

	// create resource which is already present it should update
	dr2 := fakeDr
	dr2.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr2)
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

	// create resource2
	dr3 := newFakeDr
	dr3.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr3)
	cdr3, err3 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(newFakeDiskUid, metav1.GetOptions{})

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
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	// update a resource which is not present
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	err := fakeController.UpdateDisk(dr, nil)
	if err == nil {
		t.Error("error should not be nil as the resource is not present")
	}
	// create one disk resource then update it.
	fakeController.CreateDisk(dr)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

	// update already created resource
	err = fakeController.UpdateDisk(dr, cdr1)
	if err != nil {
		t.Fatal(err)
	}
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

	// pass nil value in old resource
	err = fakeController.UpdateDisk(dr, nil)
	if err != nil {
		t.Fatal(err)
	}
	cdr3, err3 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

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

	cdr, _ := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})
	dr.ObjectMeta.Name = "disk-updated-fake-uuid"
	err = fakeController.UpdateDisk(dr, cdr)
	if err == nil {
		t.Error("if resource is not present then it should return error")
	}
}

func TestDeactivateDisk(t *testing.T) {
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	// create one resource and deactivate it.
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr)
	fakeController.DeactivateDisk(dr)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})
	// deactivate one resource which is not present it should return error
	dr1 := newFakeDr
	dr1.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.DeactivateDisk(dr1)
	// create another resource and deactivate it.
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(newDr)
	fakeController.DeactivateDisk(newDr)
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(newFakeDiskUid, metav1.GetOptions{})

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"deactivete dr resource":    {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"deactivete newDr resource": {actualDisk: *cdr2, actualError: err2, expectedDisk: newDr, expectedError: nil},
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
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	// create one resource and delete it
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr)
	fakeController.DeleteDisk(fakeDiskUid)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})
	// delete one resource which is not present it should return error
	fakeController.DeleteDisk("another-uuid")
	// create another resource and delete it
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(newDr)
	fakeController.DeleteDisk(newFakeDiskUid)
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(newFakeDiskUid, metav1.GetOptions{})

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
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	// create resource1
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr)
	// create resource2
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(newDr)
	listDevice, err := fakeController.listDiskResource()
	typeMeta := metav1.TypeMeta{}
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

func TestGetExistingResource(t *testing.T) {
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	// create resource1
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(dr)
	// create resource2
	newDr := newFakeDr
	newDr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.CreateDisk(newDr)

	cdr1 := fakeController.getExistingResource(fakeDiskUid)
	cdr2 := fakeController.getExistingResource(newFakeDiskUid)
	cdr3 := fakeController.getExistingResource("newFakeDiskUid")
	tests := map[string]struct {
		actualDisk   *apis.Disk
		expectedDisk *apis.Disk
	}{
		"resouce with 'fake-disk-uid' uuid":             {actualDisk: cdr1, expectedDisk: &dr},
		"resouce with 'new-fake-disk-uid' uuid":         {actualDisk: cdr2, expectedDisk: &newDr},
		"resouce with invalid uuid not present in etcd": {actualDisk: cdr3, expectedDisk: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}
