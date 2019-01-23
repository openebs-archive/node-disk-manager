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
	"fmt"
	"testing"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ndmFakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	fakeDiskUid     = "fake-disk-uid"
	newFakeDiskUid  = "new-fake-disk-uid"
	fakeHostName    = "fake-host-name"
	newFakeHostName = "new-fake-host-name"
)

var (
	// mock data for new disk.
	fakeCapacity = apis.DiskCapacity{
		Storage: 100000,
	}
	fakeDetails = apis.DiskDetails{
		Model:  "disk-fake-model",
		Serial: "disk-fake-serial",
		Vendor: "disk-fake-vendor",
	}
	fakeObj = apis.DiskSpec{
		Path:     "dev/disk-fake-path",
		Capacity: fakeCapacity,
		Details:  fakeDetails,
		DevLinks: make([]apis.DiskDevLink, 0),
	}
	fakeTypeMeta = metav1.TypeMeta{
		Kind:       NDMDiskKind,
		APIVersion: NDMVersion,
	}
	fakeObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   fakeDiskUid,
	}
	fakeDiskStatus = apis.DiskStatus{
		State: NDMActive,
	}
	fakeDr = apis.Disk{
		TypeMeta:   fakeTypeMeta,
		ObjectMeta: fakeObjectMeta,
		Spec:       fakeObj,
		Status:     fakeDiskStatus,
	}

	// mock data for new disk.
	newFakeCapacity = apis.DiskCapacity{
		Storage: 200000,
	}
	newFakeDetails = apis.DiskDetails{
		Model:  "disk-fake-model-new",
		Serial: "disk-fake-serial-new",
		Vendor: "disk-fake-vendor-new",
	}
	newFakeObj = apis.DiskSpec{
		Path:     "dev/disk-fake-path-new",
		Capacity: newFakeCapacity,
		Details:  newFakeDetails,
		DevLinks: make([]apis.DiskDevLink, 0),
	}
	newFakeTypeMeta = metav1.TypeMeta{
		Kind:       NDMDiskKind,
		APIVersion: NDMVersion,
	}
	newFakeObjectMeta = metav1.ObjectMeta{
		Labels: make(map[string]string),
		Name:   newFakeDiskUid,
	}
	newFakeDiskStatus = apis.DiskStatus{
		State: NDMActive,
	}
	newFakeDr = apis.Disk{
		TypeMeta:   newFakeTypeMeta,
		ObjectMeta: newFakeObjectMeta,
		Spec:       newFakeObj,
		Status:     newFakeDiskStatus,
	}
)

func CreateFakeClient(t *testing.T) client.Client {
	dr := &apis.Disk{
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

	objs := []runtime.Object{dr}
	s := scheme.Scheme
	s.AddKnownTypes(apis.SchemeGroupVersion, dr)
	s.AddKnownTypes(apis.SchemeGroupVersion, diskList)
	fakeNdmClient := ndmFakeClientset.NewFakeClient(objs...)
	if fakeNdmClient == nil {
		fmt.Println("NDMClient is not created")
	}
	return fakeNdmClient
}
