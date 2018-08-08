/*
Copyright 2018 The OpenEBS Authors.

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

package filter

import (
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/stretchr/testify/assert"
)

func TestOsDiskFilterRegister(t *testing.T) {
	diskDetails, err := libudevwrapper.MockDiskDetails()
	if err != nil {
		t.Fatal(err)
	}
	expectedFilterList := make([]*controller.Filter, 0)
	fakeController := &controller.Controller{
		Filters: make([]*controller.Filter, 0),
		Mutex:   &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	oSDiskExcludeFilterRegister()
	var fi controller.FilterInterface = &oSDiskExcludeFilter{
		controller:     fakeController,
		excludeDevPath: diskDetails.DevNode,
	}
	filter := &controller.Filter{
		Name:      oSDiskExcludeFilterName,
		State:     oSDiskExcludeFilterState,
		Interface: fi,
	}
	expectedFilterList = append(expectedFilterList, filter)
	tests := map[string]struct {
		actualFilterList   []*controller.Filter
		expectedFilterList []*controller.Filter
	}{
		"add one filter and check if it is present or not": {actualFilterList: fakeController.Filters, expectedFilterList: expectedFilterList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedFilterList, test.actualFilterList)
		})
	}
}

func TestOsDiskExcludeFilterExclude(t *testing.T) {
	fakeDiskPath := "fake-disk-path"
	ignoreDiskPath := "ignore-disk-path"
	fakeOsDiskFilter := oSDiskExcludeFilter{excludeDevPath: ignoreDiskPath}
	disk1 := &controller.DiskInfo{}
	disk1.Path = fakeDiskPath
	disk2 := &controller.DiskInfo{}
	disk2.Path = ignoreDiskPath
	tests := map[string]struct {
		disk     *controller.DiskInfo
		actual   bool
		expected bool
	}{
		"comparing return of Exclude for disk1": {actual: fakeOsDiskFilter.Exclude(disk1), expected: true},
		"comparing return of Exclude for disk2": {actual: fakeOsDiskFilter.Exclude(disk2), expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.actual)
		})
	}
}

func TestOsDiskExcludeFilterInclude(t *testing.T) {
	fakeDiskPath := "fake-disk-path"
	ignoreDiskPath := "ignore-disk-path"
	fakeOsDiskFilter := oSDiskExcludeFilter{excludeDevPath: ignoreDiskPath}
	disk1 := &controller.DiskInfo{}
	disk1.Path = fakeDiskPath
	disk2 := &controller.DiskInfo{}
	disk2.Path = ignoreDiskPath
	tests := map[string]struct {
		disk     *controller.DiskInfo
		actual   bool
		expected bool
	}{
		"comparing return of Include for disk1": {actual: fakeOsDiskFilter.Include(disk1), expected: true},
		"comparing return of Include for disk2": {actual: fakeOsDiskFilter.Include(disk2), expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.actual)
		})
	}
}
