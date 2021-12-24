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

package controller

import (
	"sort"

	"k8s.io/klog"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// EventMessage struct contains attribute of event message info.
type EventMessage struct {
	Action          string                     // Action is event action like attach/detach
	Devices         []*blockdevice.BlockDevice // list of block device details
	RequestedProbes []string                   // List of probes (given as probe names) to be run for this event. Optional
	AllBlockDevices bool                       // If true, ignore Devices list and iterate through all block devices present in the hierarchy cache.
}

var EventMessageChannel = make(chan EventMessage)

// Probe contains name, state and probeinterface
type Probe struct {
	Priority  int
	Name      string
	State     bool
	Interface ProbeInterface
}

// Start implements ProbeInterface's Start()
func (p *Probe) Start() {
	p.Interface.Start()
}

// FillBlockDeviceDetails implements ProbeInterface's FillBlockDeviceDetails()
func (p *Probe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	p.Interface.FillBlockDeviceDetails(blockDevice)
}

// ProbeInterface contains Start() and  FillBlockDeviceDetails()
type ProbeInterface interface {
	Start()
	FillBlockDeviceDetails(*blockdevice.BlockDevice)
}

// sortableProbes contains a slice of probes
type sortableProbes []*Probe

// Len returns the length of a slice.
func (ps sortableProbes) Len() int {
	return len(ps)
}

// Swap swaps the elements with indexes i and j.
func (ps sortableProbes) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ps sortableProbes) Less(i, j int) bool {
	return ps[i].Priority < ps[j].Priority
}

// AddNewProbe adds new probe to controller object
func (c *Controller) AddNewProbe(probe *Probe) {
	c.Lock()
	defer c.Unlock()
	probes := c.Probes
	probes = append(probes, probe)
	sort.Sort(sortableProbes(probes))
	c.Probes = probes
	klog.Info("configured ", probe.Name, " : state ", util.StateStatus(probe.State))
}

// ListProbe returns list of active probe associated with controller object.
// optinally pass a list of probe names to select only from the passed probes.
func (c *Controller) ListProbe(requestedProbes ...string) []*Probe {
	c.Lock()
	defer c.Unlock()
	allProbes := false
	if len(requestedProbes) == 0 {
		allProbes = true
	}
	listProbe := make([]*Probe, 0)
	for _, probe := range c.Probes {
		if probe.State && (allProbes ||
			util.Contains(requestedProbes, probe.Name)) {
			listProbe = append(listProbe, probe)
		}
	}
	return listProbe
}

// FillBlockDeviceDetails lists registered probes and fills details from each probe
func (c *Controller) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice,
	requestedProbes ...string) {
	blockDevice.NodeAttributes = c.NodeAttributes
	blockDevice.Labels = make(map[string]string)
	selectedProbes := c.ListProbe(requestedProbes...)
	for _, probe := range selectedProbes {
		probe.FillBlockDeviceDetails(blockDevice)
		klog.Info("details filled by ", probe.Name)
	}
}
