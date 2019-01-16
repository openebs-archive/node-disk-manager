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

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/stretchr/testify/assert"
)

type fakeFilter struct {
	ctrl *controller.Controller
}

func (f *fakeFilter) Start() {}

func (f *fakeFilter) Include(*controller.DiskInfo) bool {
	return false
}
func (f *fakeFilter) Exclude(*controller.DiskInfo) bool {
	return true
}

//Add one new filter and get the list of the filters and match them
func TestRegisterFilter(t *testing.T) {
	expectedFilterList := make([]*controller.Filter, 0)
	filters := make([]*controller.Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		Filters: filters,
		Mutex:   mutex,
	}
	var i controller.FilterInterface = &fakeFilter{}
	newRegisterFilter := &registerFilter{
		name:       "filter-1",
		state:      true,
		fi:         i,
		controller: fakeController,
	}
	newRegisterFilter.register()
	filter := &controller.Filter{
		Name:      newRegisterFilter.name,
		State:     newRegisterFilter.state,
		Interface: newRegisterFilter.fi,
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

func TestStart(t *testing.T) {
	expectedFilterList := make([]*controller.Filter, 0)
	fakeController := &controller.Controller{
		Filters: make([]*controller.Filter, 0),
		Mutex:   &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	var fakeFilterRegister = func() {
		ctrl := <-controller.ControllerBroadcastChannel
		if ctrl == nil {
			t.Fatal("controller struct should not be nil")
		}
		var fi controller.FilterInterface = &fakeFilter{ctrl: ctrl}
		newRegisterFilter := &registerFilter{
			name:       "fake-filter",
			state:      defaultEnabled,
			fi:         fi,
			controller: ctrl,
		}
		newRegisterFilter.register()
	}
	var registeredFilters = []func(){fakeFilterRegister}
	Start(registeredFilters)
	var fi controller.FilterInterface = &fakeFilter{ctrl: fakeController}
	filter := &controller.Filter{
		Name:      "fake-filter",
		State:     defaultEnabled,
		Interface: fi,
	}
	expectedFilterList = append(expectedFilterList, filter)
	tests := map[string]struct {
		actualFilterList   []*controller.Filter
		expectedFilterList []*controller.Filter
	}{
		"register one filter and check if it is present or not": {actualFilterList: fakeController.Filters, expectedFilterList: expectedFilterList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedFilterList, test.actualFilterList)
		})
	}
}
