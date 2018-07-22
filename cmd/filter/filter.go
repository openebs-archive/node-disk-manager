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
)

const (
	defaultEnabled  = true  // use in each filter to make it enable.
	defaultDisabled = false // use in each filter to make it disable.
)

// RegisterdFilters contains register function of filters which we want to register
var RegisterdFilters = []func(){oSDiskExludeFilterRegister}

type registerFilter struct {
	name       string
	state      bool
	fi         controller.FilterInterface
	controller *controller.Controller
}

// register called by register function of each filter it will check for filter
// status if it is enabled then it will call Start() of that filter.
func (rf *registerFilter) register() {
	newFilter := &controller.Filter{
		Name:      rf.name,
		State:     rf.state,
		Interface: rf.fi,
	}
	rf.controller.AddNewFilter(newFilter)
	if rf.state {
		rf.fi.Start()
	}
}

// Start() starts registration of filters present in RegisteredFilters
func Start(registerdFilters []func()) {
	glog.Info("registering filters")
	for _, filter := range registerdFilters {
		filter()
	}
}
