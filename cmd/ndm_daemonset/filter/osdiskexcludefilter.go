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

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	osDiskExcludeFilterKey = "os-disk-exclude-filter"
)

var (
	defaultMountFilePath     = "/proc/self/mounts"
	mountPoints              = []string{"/", "/etc/hosts"}
	hostMountFilePath        = "/host/mounts"           // hostMountFilePath is the file path mounted inside container
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
	controller     *controller.Controller
	excludeDevPath string
}

// newOsDiskFilter returns new pointer osDiskFilter
func newNonOsDiskFilter(ctrl *controller.Controller) *oSDiskExcludeFilter {
	return &oSDiskExcludeFilter{
		controller: ctrl,
	}
}

// Start set os disk devPath in nonOsDiskFilter pointer
func (odf *oSDiskExcludeFilter) Start() {
	/*
		process1 check for mentioned mountpoint in host's /proc/1/mounts file
		host's /proc/1/mounts file mounted inside container it checks for
		every mentioned mountpoints if it is able to get parent disk's devpath
		it adds it in Controller struct and make isOsDiskFilterSet true
	*/
	for _, mountPoint := range mountPoints {
		mountPointUtil := util.NewMountUtil(hostMountFilePath, mountPoint)
		if devPath, err := mountPointUtil.GetDiskPath(); err == nil {
			odf.excludeDevPath = devPath
			return
		} else {
			glog.Error(err)
		}
	}
	/*
		process2 check for mountpoints in /proc/self/mounts file if it is able to get
		disk's path of it adds it in Controller struct and make isOsDiskFilterSet true
	*/
	for _, mountPoint := range mountPoints {
		mountPointUtil := util.NewMountUtil(defaultMountFilePath, mountPoint)
		if devPath, err := mountPointUtil.GetDiskPath(); err == nil {
			odf.excludeDevPath = devPath
			return
		} else {
			glog.Error(err)
		}
	}
	glog.Error("unable to apply os disk filter")
}

// Include contains nothing by default it returns false
func (odf *oSDiskExcludeFilter) Include(d *controller.DiskInfo) bool {
	return true
}

// Exclude returns true if disk devpath does not match with excludeDevPath
func (odf *oSDiskExcludeFilter) Exclude(d *controller.DiskInfo) bool {
	// The partitionRegex is chosen depending on whether the device uses
	// the p[0-9] partition naming structure or not.
	var partitionRegex string
	if util.MatchRegex(".+[0-9]+$", odf.excludeDevPath) {
		// matches loop0, loop0p1, nvme3n0p1
		partitionRegex = "(p[0-9]+)?$"
	} else {
		// matches sda, sda1
		partitionRegex = "[0-9]*$"
	}
	glog.Info("applying regex ", odf.excludeDevPath+partitionRegex, " for os-disk-exclude-filter")
	return !util.MatchRegex(odf.excludeDevPath+partitionRegex, d.Path)
}
