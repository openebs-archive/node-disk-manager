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
	"k8s.io/klog"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// Filter contains name, state and filterInterface
type Filter struct {
	Name      string          // Name is the name of the filter
	State     bool            // State is the State of the filter
	Interface FilterInterface // Interface contains registered filter
}

// ApplyFilter returns true if both any of include() or exclude() returns true.
// We are having two types of filter function one is inclusion and exclusion type
// if any of them returns true then filter doesn't want further process of that event.
func (f *Filter) ApplyFilter(diskInfo *DiskInfo) bool {
	return f.Interface.Include(diskInfo) && f.Interface.Exclude(diskInfo)

}

// Start implements FilterInterface's Start()
func (f *Filter) Start() {
	f.Interface.Start()
}

// FilterInterface contains Filters interface and Start()
type FilterInterface interface {
	Start()
	Filters
}

// Filters contains Include() and Exclude() filter method. There
// will be some preset value of include and exclude if passing DiskInfo
// matches with include value then it returns true if passing DiskINfo
// does not match with exclude value then it returns false
type Filters interface {
	Include(*DiskInfo) bool // Include returns True if passing DiskInfo matches with include value
	Exclude(*DiskInfo) bool // exclude returns True if passing DiskInfo does not matche with exclude value
}

// AddNewFilter adds new filter to controller object
func (c *Controller) AddNewFilter(filter *Filter) {
	c.Lock()
	defer c.Unlock()
	filters := c.Filters
	filters = append(filters, filter)
	c.Filters = filters
	klog.Info("configured ", filter.Name, " : state ", util.StateStatus(filter.State))
}

// ListFilter returns list of active filters associated with controller object
func (c *Controller) ListFilter() []*Filter {
	c.Lock()
	defer c.Unlock()
	listFilter := make([]*Filter, 0)
	for _, filter := range c.Filters {
		if filter.State {
			listFilter = append(listFilter, filter)
		}
	}
	return listFilter
}

// ApplyFilter checks status for every registered filters if any of the filters
// wants to stop further process of the event it returns true else it returns false
func (c *Controller) ApplyFilter(diskDetails *DiskInfo) bool {
	for _, filter := range c.ListFilter() {
		if !filter.ApplyFilter(diskDetails) {
			klog.Info(diskDetails.Uuid, " ignored by ", filter.Name)
			return false
		}
	}
	return true
}
