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

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	vendorFilterKey = "vendor-filter"
)

var (
	vendorFilterName  = "vendor filter" // filter name
	vendorFilterState = defaultEnabled  // filter state
	includeVendors    = ""
	excludeVendors    = ""
)

// vendorFilterRegister contains registration process of VendorFilter
var vendorFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}
	if ctrl.NDMConfig != nil {
		for _, filterConfig := range ctrl.NDMConfig.FilterConfigs {
			if filterConfig.Key == vendorFilterKey {
				vendorFilterName = filterConfig.Name
				vendorFilterState = util.CheckTruthy(filterConfig.State)
				includeVendors = filterConfig.Include
				excludeVendors = filterConfig.Exclude
				break
			}
		}
	}
	var fi controller.FilterInterface = newVendorFilter(ctrl)
	newRegisterFilter := &registerFilter{
		name:       vendorFilterName,
		state:      vendorFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newRegisterFilter.register()
}

// vendorFilter contains controller and include and exclude vendors
type vendorFilter struct {
	controller     *controller.Controller
	excludeVendors []string
	includeVendors []string
}

// newVendorFilter returns new pointer osDiskFilter
func newVendorFilter(ctrl *controller.Controller) *vendorFilter {
	return &vendorFilter{
		controller: ctrl,
	}
}

// Start sets include and exclude vendor's list
func (vf *vendorFilter) Start() {
	vf.includeVendors = make([]string, 0)
	vf.excludeVendors = make([]string, 0)
	if includeVendors != "" {
		vf.includeVendors = strings.Split(includeVendors, ",")
	}
	if excludeVendors != "" {
		vf.excludeVendors = strings.Split(excludeVendors, ",")
	}
}

// Include returns true if vendor of the disk matches with given list
// or the list of the length is 0
func (vf *vendorFilter) Include(blockDevice *blockdevice.BlockDevice) bool {
	if len(vf.includeVendors) == 0 {
		return true
	}
	return util.ContainsIgnoredCase(vf.includeVendors, blockDevice.DeviceDetails.Vendor)
}

// Exclude returns true if vendor of the disk does not match with given
// list or the list of the length is 0
func (vf *vendorFilter) Exclude(blockDevice *blockdevice.BlockDevice) bool {
	if len(vf.excludeVendors) == 0 {
		return true
	}
	return !util.ContainsIgnoredCase(vf.excludeVendors, blockDevice.DeviceDetails.Vendor)
}
