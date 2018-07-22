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
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	oSDiskExludeFilterName  = "os disk exlude filter" // filter name
	oSDiskExludeFilterState = defaultEnabled          // filter state
)

var (
	defaultMountFilePath = "/proc/self/mounts"
	mountPoints          = []string{"/", "/etc/hosts"}
	hostMountFilePath    = "/host/mounts" // hostMountFilePath is the file path mounted inside container
)

// oSDiskExludeFilterRegister contains registration process of oSDiskExludeFilter
var oSDiskExludeFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}
	var fi controller.FilterInterface = newNonOsDiskFilter(ctrl)
	newPrgisterFilter := &registerFilter{
		name:       oSDiskExludeFilterName,
		state:      oSDiskExludeFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newPrgisterFilter.register()
}

// oSDiskExludeFilter controller and path of os disk
type oSDiskExludeFilter struct {
	controller     *controller.Controller
	excludeDevPath string
}

// newOsDiskFilter returns new pointer osDiskFilter
func newNonOsDiskFilter(ctrl *controller.Controller) *oSDiskExludeFilter {
	return &oSDiskExludeFilter{
		controller: ctrl,
	}
}

// Start set os disk devPath in nonOsDiskFilter pointer
func (odf *oSDiskExludeFilter) Start() {
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
		}
	}
	glog.Error("unable to apply os disk filter")
}

// Include contains nothing by default it returns false
func (odf *oSDiskExludeFilter) Include(d *controller.DiskInfo) bool {
	return false
}

// Exclude returns true if disk devpath matches with excludeDevPath
func (odf *oSDiskExludeFilter) Exclude(d *controller.DiskInfo) bool {
	return odf.excludeDevPath != d.Path
}
