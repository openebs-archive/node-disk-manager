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
	"github.com/openebs/node-disk-manager/blockdevice"
	"strings"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	pathFilterKey = "path-filter"
)

var (
	pathFilterName  = "path filter"  // filter device paths
	pathFilterState = defaultEnabled // filter state
	includePaths    = ""
	excludePaths    = "loop"
)

// pathFilterRegister contains registration process of PathFilter
var pathFilterRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		return
	}
	if ctrl.NDMConfig != nil {
		for _, filterConfig := range ctrl.NDMConfig.FilterConfigs {
			if filterConfig.Key == pathFilterKey {
				pathFilterName = filterConfig.Name
				pathFilterState = util.CheckTruthy(filterConfig.State)
				includePaths = filterConfig.Include
				excludePaths = filterConfig.Exclude
				break
			}
		}
	}
	var fi controller.FilterInterface = newPathFilter(ctrl)
	newRegisterFilter := &registerFilter{
		name:       pathFilterName,
		state:      pathFilterState,
		fi:         fi,
		controller: ctrl,
	}
	newRegisterFilter.register()
}

// pathFilter contains controller and include and exclude keywords
type pathFilter struct {
	controller   *controller.Controller
	excludePaths []string
	includePaths []string
}

// newPathFilter returns new pointer PathFilter
func newPathFilter(ctrl *controller.Controller) *pathFilter {
	return &pathFilter{
		controller: ctrl,
	}
}

// Start sets include and exclude path keywords list
func (pf *pathFilter) Start() {
	pf.includePaths = make([]string, 0)
	pf.excludePaths = make([]string, 0)
	if includePaths != "" {
		pf.includePaths = strings.Split(includePaths, ",")
	}
	if excludePaths != "" {
		pf.excludePaths = strings.Split(excludePaths, ",")
	}
}

// Include returns true if the disk path matches with given list
// of keywords
func (pf *pathFilter) Include(blockDevice *blockdevice.BlockDevice) bool {
	if len(pf.includePaths) == 0 {
		return true
	}
	return util.MatchIgnoredCase(pf.includePaths, blockDevice.DevPath)
}

// Exclude returns true if the disk path does not match any given
// keywords
func (pf *pathFilter) Exclude(blockDevice *blockdevice.BlockDevice) bool {
	if len(pf.excludePaths) == 0 {
		return true
	}
	return !util.MatchIgnoredCase(pf.excludePaths, blockDevice.DevPath)
}
