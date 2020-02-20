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
	"strings"
	"sync"
	"testing"

	. "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"

	"github.com/stretchr/testify/assert"
)

func TestVendorFilterRegister(t *testing.T) {
	expectedFilterList := make([]*controller.Filter, 0)
	fakeController := &controller.Controller{
		Filters: make([]*controller.Filter, 0),
		Mutex:   &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	vendorFilterRegister()
	var fi controller.FilterInterface = &vendorFilter{
		controller:     fakeController,
		includeVendors: make([]string, 0),
		excludeVendors: make([]string, 0),
	}
	filter := &controller.Filter{
		Name:      vendorFilterName,
		State:     vendorFilterState,
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

func TestVendorStart(t *testing.T) {
	fakeVendorFilter1 := vendorFilter{}
	fakeVendorFilter2 := vendorFilter{}
	tests := map[string]struct {
		filter        vendorFilter
		includeVendor string
		excludeVendor string
	}{
		"includeVendor is empty":           {filter: fakeVendorFilter1, includeVendor: "", excludeVendor: ""},
		"includeVendor and vendor is same": {filter: fakeVendorFilter2, includeVendor: "Google", excludeVendor: "Google"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			includeVendors = test.includeVendor
			excludeVendors = test.excludeVendor
			test.filter.Start()
			if test.excludeVendor != "" {
				assert.Equal(t, strings.Split(test.excludeVendor, ","), test.filter.excludeVendors)
			} else {
				assert.Equal(t, make([]string, 0), test.filter.excludeVendors)
			}
			if test.includeVendor != "" {
				assert.Equal(t, strings.Split(test.excludeVendor, ","), test.filter.includeVendors)
			} else {
				assert.Equal(t, make([]string, 0), test.filter.includeVendors)
			}
		})
	}
}

func TestVendorFilterExclude(t *testing.T) {
	fakeVendorFilter1 := vendorFilter{}
	fakeVendorFilter2 := vendorFilter{}
	fakeVendorFilter3 := vendorFilter{}
	tests := map[string]struct {
		filter        vendorFilter
		excludeVendor string
		vendor        string
		expected      bool
	}{
		"excludeVendor is empty":               {filter: fakeVendorFilter1, excludeVendor: "", vendor: "SanDisk", expected: true},
		"excludeVendor and vendor is same":     {filter: fakeVendorFilter2, excludeVendor: "Google", vendor: "Google", expected: false},
		"excludeVendor and vendor is not same": {filter: fakeVendorFilter3, excludeVendor: "Google", vendor: "SanDisk", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bd := &BlockDevice{}
			bd.DeviceDetails.Vendor = test.vendor
			if test.excludeVendor != "" {
				test.filter.excludeVendors = strings.Split(test.excludeVendor, ",")
			}
			assert.Equal(t, test.expected, test.filter.Exclude(bd))
		})
	}
}

func TestVendorFilterInclude(t *testing.T) {
	fakeVendorFilter1 := vendorFilter{}
	fakeVendorFilter2 := vendorFilter{}
	fakeVendorFilter3 := vendorFilter{}
	tests := map[string]struct {
		filter        vendorFilter
		includeVendor string
		vendor        string
		expected      bool
	}{
		"includeVendor is empty":               {filter: fakeVendorFilter1, includeVendor: "", vendor: "SanDisk", expected: true},
		"includeVendor and vendor is same":     {filter: fakeVendorFilter2, includeVendor: "Google", vendor: "Google", expected: true},
		"includeVendor and vendor is not same": {filter: fakeVendorFilter3, includeVendor: "Google", vendor: "SanDisk", expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bd := &BlockDevice{}
			bd.DeviceDetails.Vendor = test.vendor
			if test.includeVendor != "" {
				test.filter.includeVendors = strings.Split(test.includeVendor, ",")
			}
			assert.Equal(t, test.expected, test.filter.Include(bd))
		})
	}
}
