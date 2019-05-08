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
	"github.com/openebs/node-disk-manager/pkg/udev"
	"strings"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	osDiskExcludeFilterKey = "os-disk-exclude-filter"
)

var (
	excludeMountPoints       = "/"
	oSDiskExcludeFilterName  = "os disk exclude filter" // filter name
	oSDiskExcludeFilterState = defaultEnabled           // filter state
)

// oSDiskExcludeFilterRegister contains registration process of oSDiskExcludeFilter
var oSDiskExcludeFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}
	if ctrl.NDMConfig != nil {
		for _, filterConfig := range ctrl.NDMConfig.FilterConfigs {
			if filterConfig.Key == osDiskExcludeFilterKey {
				oSDiskExcludeFilterName = filterConfig.Name
				oSDiskExcludeFilterState = util.CheckTruthy(filterConfig.State)
				excludeMountPoints = filterConfig.Exclude
				break
			}
		}
	}
	var fi controller.FilterInterface = newNonOsDiskFilter(ctrl)
	newRegisterFilter := &registerFilter{
		name:       oSDiskExcludeFilterName,
		state:      oSDiskExcludeFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newRegisterFilter.register()
}

// oSDiskExcludeFilter controller and path of os disk
type oSDiskExcludeFilter struct {
	controller         *controller.Controller
	excludeMountPoints []string
}

// newOsDiskFilter returns new pointer osDiskFilter
func newNonOsDiskFilter(ctrl *controller.Controller) *oSDiskExcludeFilter {
	return &oSDiskExcludeFilter{
		controller: ctrl,
	}
}

// Start set os disk devPath in nonOsDiskFilter pointer
func (odf *oSDiskExcludeFilter) Start() {
	odf.excludeMountPoints = make([]string, 0)
	if excludeMountPoints != "" {
		odf.excludeMountPoints = strings.Split(excludeMountPoints, ",")
	}

}

// Include contains nothing by default it returns false
func (odf *oSDiskExcludeFilter) Include(d *controller.DiskInfo) bool {
	return true
}

// Exclude returns true if disk mountPoint doesn't match with excludeMountPoints
func (odf *oSDiskExcludeFilter) Exclude(d *controller.DiskInfo) bool {
	if len(odf.excludeMountPoints) == 0 {
		return true
	}
	if d.DiskType == udev.UDEV_SYSTEM {
		return !util.MatchIgnoredCase(odf.excludeMountPoints, d.FileSystemInformation.MountPoint)
	} else if d.DiskType == udev.UDEV_PARTITION {
		return !util.MatchIgnoredCase(odf.excludeMountPoints, d.PartitionData[0].FileSystemInformation.MountPoint)
	}
	return true
}
