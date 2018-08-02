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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var matchDiskUuid = "fake-match-uuid"

type fakeFilter struct{}

func (f *fakeFilter) Start() {
	messageChannel <- message
}

func (f *fakeFilter) Include(fakeDiskInfo *DiskInfo) bool {
	return true
}

func (f *fakeFilter) Exclude(fakeDiskInfo *DiskInfo) bool {
	return fakeDiskInfo.Uuid != matchDiskUuid
}

//Add one new filter and get the list of the filters and match them
func TestAddNewFilter(t *testing.T) {
	filters := make([]*Filter, 0)
	expectedFilterList := make([]*Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Filters: filters,
		Mutex:   mutex,
	}
	filter := &fakeFilter{}
	filter1 := &Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}
	fakeController.AddNewFilter(filter1)
	expectedFilterList = append(expectedFilterList, filter1)
	tests := map[string]struct {
		actualFilterList   []*Filter
		expectedFilterList []*Filter
	}{
		"add one filter and check if it is present or not": {actualFilterList: fakeController.Filters, expectedFilterList: expectedFilterList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedFilterList, test.actualFilterList)
		})
	}
}

//Add some new filters and get the list of the filters and match them
func TestListFilter(t *testing.T) {
	filters := make([]*Filter, 0)
	expectedFilterList := make([]*Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Filters: filters,
		Mutex:   mutex,
	}
	filter := &fakeFilter{}
	filter1 := &Filter{
		Name:      "probe1",
		State:     true,
		Interface: filter,
	}
	filter2 := &Filter{
		Name:      "filter2",
		State:     true,
		Interface: filter,
	}
	fakeController.AddNewFilter(filter1)
	fakeController.AddNewFilter(filter2)
	expectedFilterList = append(expectedFilterList, filter1)
	expectedFilterList = append(expectedFilterList, filter2)
	tests := map[string]struct {
		actualFilterList   []*Filter
		expectedFilterList []*Filter
	}{
		"add some filters and check if they are present or not": {actualFilterList: fakeController.ListFilter(), expectedFilterList: expectedFilterList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedFilterList, test.actualFilterList)
		})
	}
}

func TestSrartFilter(t *testing.T) {
	var msg1 string
	filter := &fakeFilter{}
	filter1 := &Filter{
		Name:      "filter1",
		State:     true,
		Interface: filter,
	}
	go filter1.Start()
	select {
	case res := <-messageChannel:
		msg1 = res
	case <-time.After(1 * time.Second):
		msg1 = ""
	}

	tests := map[string]struct {
		actualMessage   string
		expectedMessage string
	}{
		"comparing message from start method of filter": {actualMessage: msg1, expectedMessage: message},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedMessage, test.actualMessage)
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	fakeFilter := &fakeFilter{}
	filter := &Filter{
		Name:      "filter1",
		State:     true,
		Interface: fakeFilter,
	}
	disk := &DiskInfo{}
	tests := map[string]struct {
		actual   bool
		expected bool
	}{
		"comparing return of ApplyFilter": {actual: filter.ApplyFilter(disk), expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.actual)
		})
	}
}

func TestApplyFilter(t *testing.T) {
	filters := make([]*Filter, 0)
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		Filters: filters,
		Mutex:   mutex,
	}
	fakeFilter := &fakeFilter{}
	filter := &Filter{
		Name:      "probe1",
		State:     true,
		Interface: fakeFilter,
	}
	fakeController.AddNewFilter(filter)
	disk1 := &DiskInfo{}
	disk2 := &DiskInfo{}
	disk2.Uuid = matchDiskUuid
	tests := map[string]struct {
		disk     *DiskInfo
		actual   bool
		expected bool
	}{
		"comparing return of ApplyFilter for disk1": {actual: fakeController.ApplyFilter(disk1), expected: true},
		"comparing return of ApplyFilter for disk2": {actual: fakeController.ApplyFilter(disk2), expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.actual)
		})
	}
}
