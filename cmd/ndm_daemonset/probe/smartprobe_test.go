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
<<<<<<< HEAD
	"github.com/openebs/node-disk-manager/blockdevice"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
=======
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"

	apis "github.com/openebs/node-disk-manager/api/v1alpha1"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
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
		LogicalBlockSize: mockOsDiskDetails.LBSize,
	}

	fakeCapacity := apis.DeviceCapacity{
		Storage:           mockOsDiskDetails.Capacity,
		LogicalSectorSize: mockOsDiskDetails.LBSize,
	}

<<<<<<< HEAD
	fakeObj := apis.DeviceSpec{
=======
	fakeObj := apis.BlockDeviceSpec{
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
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

<<<<<<< HEAD
	fakeDiskStatus := apis.DeviceStatus{
=======
	fakeDiskStatus := apis.BlockDeviceStatus{
>>>>>>> 3bfc5e1e... Inital project structuring and adding BlockDevice type
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
	expectedDiskInfo.DeviceAttributes.LogicalBlockSize = mockOsDiskDetails.LBSize
	expectedDiskInfo.DeviceAttributes.FirmwareRevision = mockOsDiskDetails.FirmwareRevision
	expectedDiskInfo.DeviceAttributes.Compliance = mockOsDiskDetails.Compliance
	assert.Equal(t, expectedDiskInfo, actualDiskInfo)
}
