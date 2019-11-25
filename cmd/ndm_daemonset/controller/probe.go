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

	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

// EventMessage struct contains attribute of event message info.
type EventMessage struct {
	Action  string      // Action is event action like attach/detach
	Devices []*DiskInfo // list of disks details
}

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

// FillDiskDetails implements ProbeInterface's FillDiskDetails()
func (p *Probe) FillDiskDetails(diskInfo *DiskInfo) {
	p.Interface.FillDiskDetails(diskInfo)
}

// ProbeInterface contains Start() and  FillDiskDetails()
type ProbeInterface interface {
	Start()
	FillDiskDetails(*DiskInfo)
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

// ListProbe returns list of active probe associated with controller object
func (c *Controller) ListProbe() []*Probe {
	c.Lock()
	defer c.Unlock()
	listProbe := make([]*Probe, 0)
	for _, probe := range c.Probes {
		if probe.State {
			listProbe = append(listProbe, probe)
		}
	}
	return listProbe
}

// FillDiskDetails lists registered probes and fills details from each probe
func (c *Controller) FillDiskDetails(diskDetails *DiskInfo) {
	diskDetails.NodeAttributes = c.NodeAttributes
	diskDetails.DiskType = NDMDefaultDiskType
	diskDetails.Uuid = diskDetails.ProbeIdentifiers.Uuid
	probes := c.ListProbe()
	for _, probe := range probes {
		probe.FillDiskDetails(diskDetails)
		klog.Info("details filled by ", probe.Name)
	}
}
