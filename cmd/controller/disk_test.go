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

	"github.com/stretchr/testify/assert"
)

func TestToDisk(t *testing.T) {
	fakeDiskInfo := NewDiskInfo()
	fakeDiskInfo.HostName = fakeHostName
	fakeDiskInfo.Uuid = fakeObjectMeta.Name
	fakeDiskInfo.Capacity = fakeCapacity.Storage
	fakeDiskInfo.Model = fakeDetails.Model
	fakeDiskInfo.Serial = fakeDetails.Serial
	fakeDiskInfo.Vendor = fakeDetails.Vendor
	fakeDiskInfo.Path = fakeObj.Path
	expectedDisk := fakeDr
	expectedDisk.ObjectMeta.Labels[NDMHostKey] = fakeHostName
	assert.Equal(t, expectedDisk, fakeDiskInfo.ToDisk())
}
