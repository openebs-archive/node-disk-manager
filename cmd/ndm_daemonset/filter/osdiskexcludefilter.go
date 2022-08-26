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
	"github.com/openebs/node-disk-manager/pkg/mount"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

const (
	osDiskExcludeFilterKey = "os-disk-exclude-filter"
)

var (
	defaultMountFilePath     = "/proc/self/mounts"
	mountPoints              = []string{"/", "/etc/hosts"}
	hostMountFilePath        = "/host/proc/1/mounts"    // hostMountFilePath is the file path mounted inside container
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
				mountPoints = strings.Split(filterConfig.Exclude, ",")
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
	controller      *controller.Controller
	excludeDevPaths []string
}

// newOsDiskFilter returns new pointer osDiskFilter
func newNonOsDiskFilter(ctrl *controller.Controller) *oSDiskExcludeFilter {
	return &oSDiskExcludeFilter{
		controller: ctrl,
	}
}

// Start set os disk devPath in nonOsDiskFilter pointer
func (odf *oSDiskExcludeFilter) Start() {
	for _, mountPoint := range mountPoints {
		var err error
		var devPath string

		// Check for mountpoints in both:
		//    the host's /proc/1/mounts file
		//    the /proc/self/mounts file
		// If it is found in either one and we are able to get the
		// disk's devpath, add it to the Controller struct.  Otherwise
		// log an error.

		mountPointUtil := mount.NewMountUtil(hostMountFilePath, "", mountPoint)
		if devPath, err = mountPointUtil.GetDiskPath(); err == nil {
			odf.excludeDevPaths = append(odf.excludeDevPaths, devPath)
			continue
		}

		mountPointUtil = mount.NewMountUtil(defaultMountFilePath, "", mountPoint)
		if devPath, err = mountPointUtil.GetDiskPath(); err == nil {
			odf.excludeDevPaths = append(odf.excludeDevPaths, devPath)
			continue
		}

		klog.Errorf("unable to configure os disk filter for mountpoint: %s, error: %v", mountPoint, err)
	}
}

// Include contains nothing by default it returns false
func (odf *oSDiskExcludeFilter) Include(blockDevice *blockdevice.BlockDevice) bool {
	return true
}

// Exclude returns true if disk devpath does not match with excludeDevPaths
func (odf *oSDiskExcludeFilter) Exclude(blockDevice *blockdevice.BlockDevice) bool {
	// The partitionRegex is chosen depending on whether the device uses
	// the p[0-9] partition naming structure or not.
	var partitionRegex string
	for _, excludeDevPath := range odf.excludeDevPaths {
		if util.IsMatchRegex(".+[0-9]+$", excludeDevPath) {
			// matches loop0, loop0p1, nvme3n0p1
			partitionRegex = "(p[0-9]+)?$"
		} else {
			// matches sda, sda1
			partitionRegex = "[0-9]*$"
		}
		regex := "^" + excludeDevPath + partitionRegex
		klog.Infof("applying os-filter regex %s on %s", regex, blockDevice.DevPath)
		if util.IsMatchRegex(regex, blockDevice.DevPath) {
			return false
		}
	}
	return true
}
