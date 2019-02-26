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
	"sync"
	"testing"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/stretchr/testify/assert"
)

type fakeProbe struct {
	ctrl *controller.Controller
}

func (p *fakeProbe) Start() {}

func (p *fakeProbe) FillDiskDetails(fakeDiskInfo *controller.DiskInfo) {
	fakeDiskInfo.Model = fakeModel
	fakeDiskInfo.Serial = fakeSerial
	fakeDiskInfo.Vendor = fakeVendor
}

func TestRegisterProbe(t *testing.T) {
	expectedProbeList := make([]*controller.Probe, 0)
	fakeController := &controller.Controller{
		Probes: make([]*controller.Probe, 0),
		Mutex:  &sync.Mutex{},
	}

	var i controller.ProbeInterface = &fakeProbe{}
	newRegisterProbe := &registerProbe{
		name:       "probe-1",
		state:      true,
		pi:         i,
		controller: fakeController,
	}
	newRegisterProbe.register()
	probe := &controller.Probe{
		Name:      newRegisterProbe.name,
		State:     newRegisterProbe.state,
		Interface: newRegisterProbe.pi,
	}
	expectedProbeList = append(expectedProbeList, probe)
	tests := map[string]struct {
		actualProbeList   []*controller.Probe
		expectedProbeList []*controller.Probe
	}{
		"add one probe and check if it is present or not": {actualProbeList: fakeController.Probes, expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

func TestStart(t *testing.T) {
	expectedProbeList := make([]*controller.Probe, 0)
	fakeController := &controller.Controller{
		Probes: make([]*controller.Probe, 0),
		Mutex:  &sync.Mutex{},
	}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	var fakeProbeRegister = func() {
		ctrl := <-controller.ControllerBroadcastChannel
		if ctrl == nil {
			t.Fatal("controller struct should not be nil")
		}
		var pi controller.ProbeInterface = &fakeProbe{ctrl: ctrl}
		newRegisterProbe := &registerProbe{
			name:       "fake-probe",
			state:      defaultEnabled,
			pi:         pi,
			controller: ctrl,
		}
		newRegisterProbe.register()
	}
	var registeredProbes = []func(){fakeProbeRegister}
	Start(registeredProbes)
	var fi controller.ProbeInterface = &fakeProbe{ctrl: fakeController}
	probe := &controller.Probe{
		Name:      "fake-probe",
		State:     defaultEnabled,
		Interface: fi,
	}
	expectedProbeList = append(expectedProbeList, probe)
	tests := map[string]struct {
		actualProbeList   []*controller.Probe
		expectedProbeList []*controller.Probe
	}{
		"register one probe and check if it is present or not": {actualProbeList: fakeController.Probes, expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}
