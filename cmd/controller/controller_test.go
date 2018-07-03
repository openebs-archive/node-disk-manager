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

func TestPushDiskResource(t *testing.T) {
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
	}
	dr := fakeDr
	dr.ObjectMeta.Labels[NDMHostKey] = fakeController.HostName
	fakeController.PushDiskResource(fakeDiskUid, dr)
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})
	fakeController.PushDiskResource(fakeDiskUid, dr)
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})

	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"push disk resource creates resource": {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"push disk resource updates resource": {actualDisk: *cdr2, actualError: err2, expectedDisk: dr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestDeactivateStaleDiskResource(t *testing.T) {
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
	//add one resource's uuid so state of the other resource should be inactive.
	diskList := make([]string, 0)
	diskList = append(diskList, newFakeDiskUid)
	fakeController.DeactivateStaleDiskResource(diskList)
	dr.Status.State = NDMInactive
	cdr1, err1 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(fakeDiskUid, metav1.GetOptions{})
	cdr2, err2 := fakeController.Clientset.OpenebsV1alpha1().Disks().Get(newFakeDiskUid, metav1.GetOptions{})
	tests := map[string]struct {
		actualDisk    apis.Disk
		actualError   error
		expectedDisk  apis.Disk
		expectedError error
	}{
		"resource1 present in etcd but not in system": {actualDisk: *cdr1, actualError: err1, expectedDisk: dr, expectedError: nil},
		"resource2 present in both etcd and systeme":  {actualDisk: *cdr2, actualError: err2, expectedDisk: newDr, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}
