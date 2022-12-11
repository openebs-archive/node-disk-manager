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
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"k8s.io/klog/v2"
)

const (
	defaultEnabled  = true  // use in each probe to make it enable.
	defaultDisabled = false // use in each probe to make it disable.
)

// RegisteredProbes contains register function of probes which we want to register
var RegisteredProbes = []func(){
	seachestProbeRegister,
	smartProbeRegister,
	mountProbeRegister,
	udevProbeRegister,
	sysfsProbeRegister,
	usedbyProbeRegister,
	customTagProbeRegister,
	blkidProbeRegister,
}

type registerProbe struct {
	priority   int
	name       string
	state      bool
	pi         controller.ProbeInterface
	controller *controller.Controller
}

// register called by register function of each probe it will check for probe
// status if it is enabled then it will call Start() of that probe.
func (rp *registerProbe) register() {
	newProbe := &controller.Probe{
		Priority:  rp.priority,
		Name:      rp.name,
		State:     rp.state,
		Interface: rp.pi,
	}
	rp.controller.AddNewProbe(newProbe)
	if rp.state {
		rp.pi.Start()
	}
}

// Start starts registration of probes present in RegisteredProbes
func Start(registeredProbes []func()) {
	klog.Info("registering probes")
	for _, probe := range registeredProbes {
		probe()
	}
}
