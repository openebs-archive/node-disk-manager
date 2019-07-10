/*
Copyright 2019 The OpenEBS Authors

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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiskToDeviceUUID(t *testing.T) {
	ctrl := Controller{}
	diskUID := "disk-adb1b23f7ba395988d029d78ef7bda58"
	blockDeviceUID := "blockdevice-adb1b23f7ba395988d029d78ef7bda58"
	assert.Equal(t, blockDeviceUID, ctrl.DiskToDeviceUUID(diskUID))
}
