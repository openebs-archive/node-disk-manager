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

package probe

import (
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/controller"
)

const (
	defaultEnabled  = true  // use in each probe to make it enable.
	defaultDisabled = false // use in each probe to make it disable.
)

type registerProbe struct {
	priority       int
	probeName      string
	probeState     bool
	probeInterface controller.ProbeInterface
	controller     *controller.Controller
}

// registerProbe called by init() of each probe it will check for probe status
// if it is enabled then it will call Start() of that probe.
func (rp *registerProbe) register() {
	newProbe := &controller.Probe{
		Priority:   rp.priority,
		ProbeName:  rp.probeName,
		ProbeState: rp.probeState,
		Interface:  rp.probeInterface,
	}
	rp.controller.AddNewProbe(newProbe)
	if rp.probeState {
		rp.probeInterface.Start()
	}
}

// Start() is called to invoke all init functions of each specific probe.
func Start() {
	glog.Info("starting probe")
}
