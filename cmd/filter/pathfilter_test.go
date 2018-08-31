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

	"github.com/openebs/node-disk-manager/cmd/controller"
	"github.com/stretchr/testify/assert"
)

func TestPathFilterRegister(t *testing.T) {
	expectedFilterList := make([]*controller.Filter, 0)
	fakeController := &controller.Controller{
		Filters: make([]*controller.Filter, 0),
		Mutex:   &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	pathFilterRegister()
	var fi controller.FilterInterface = &pathFilter{
		controller:   fakeController,
		includePaths: make([]string, 0),
		excludePaths: []string{"loop"},
	}
	filter := &controller.Filter{
		Name:      pathFilterName,
		State:     pathFilterState,
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

func TestPathStart(t *testing.T) {
	fakePathFilter1 := pathFilter{}
	fakePathFilter2 := pathFilter{}
	tests := map[string]struct {
		filter      pathFilter
		includePath string
		excludePath string
	}{
		"includePath is empty":         {filter: fakePathFilter1, includePath: "", excludePath: ""},
		"includePath and path is same": {filter: fakePathFilter2, includePath: "loop", excludePath: "loop"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			includePaths = test.includePath
			excludePaths = test.excludePath
			test.filter.Start()
			if test.excludePath != "" {
				assert.Equal(t, strings.Split(test.excludePath, ","), test.filter.excludePaths)
			} else {
				assert.Equal(t, make([]string, 0), test.filter.excludePaths)
			}
			if test.includePath != "" {
				assert.Equal(t, strings.Split(test.excludePath, ","), test.filter.includePaths)
			} else {
				assert.Equal(t, make([]string, 0), test.filter.includePaths)
			}
		})
	}
}

func TestPathFilterExclude(t *testing.T) {
	fakePathFilter1 := pathFilter{}
	fakePathFilter2 := pathFilter{}
	fakePathFilter3 := pathFilter{}
	tests := map[string]struct {
		filter      pathFilter
		excludePath string
		path        string
		expected    bool
	}{
		"excludePath is empty":             {filter: fakePathFilter1, excludePath: "", path: "/dev/loop0", expected: true},
		"excludePath and path is same":     {filter: fakePathFilter2, excludePath: "loop", path: "/dev/loop0", expected: false},
		"excludePath and path is not same": {filter: fakePathFilter3, excludePath: "loop", path: "/dev/sdb", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			disk := &controller.DiskInfo{}
			disk.Path = test.path
			if test.excludePath != "" {
				test.filter.excludePaths = strings.Split(test.excludePath, ",")
			}
			assert.Equal(t, test.expected, test.filter.Exclude(disk))
		})
	}
}

func TestPathFilterInclude(t *testing.T) {
	fakePathFilter1 := pathFilter{}
	fakePathFilter2 := pathFilter{}
	fakePathFilter3 := pathFilter{}
	tests := map[string]struct {
		filter      pathFilter
		includePath string
		path        string
		expected    bool
	}{
		"includePath is empty":             {filter: fakePathFilter1, includePath: "", path: "/dev/loop0", expected: true},
		"includePath and path is same":     {filter: fakePathFilter2, includePath: "loop", path: "/dev/loop0", expected: true},
		"includePath and path is not same": {filter: fakePathFilter3, includePath: "loop", path: "/dev/sdb", expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			disk := &controller.DiskInfo{}
			disk.Path = test.path
			if test.includePath != "" {
				test.filter.includePaths = strings.Split(test.includePath, ",")
			}
			assert.Equal(t, test.expected, test.filter.Include(disk))
		})
	}
}
