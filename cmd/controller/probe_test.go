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

package controller

import (
	"sync"
	"testing"
	"time"

	ndmFakeClientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

var messageChannel = make(chan string)
var message = "This is a message from start method"

type newProbe struct{}

func (np *newProbe) Start() {
	messageChannel <- message
}

func (np *newProbe) FillDiskDetails(fakeDiskInfo *DiskInfo) {
	fakeDiskInfo.HostName = fakeHostName
	fakeDiskInfo.Uuid = fakeObjectMeta.Name
	fakeDiskInfo.Capacity = fakeCapacity.Storage
	fakeDiskInfo.Model = fakeDetails.Model
	fakeDiskInfo.Serial = fakeDetails.Serial
	fakeDiskInfo.Vendor = fakeDetails.Vendor
	fakeDiskInfo.Path = fakeObj.Path
}

func TestAddNewProbe(t *testing.T) {
	probes := make([]*Probe, 0)
	expectedProbeList := make([]*Probe, 0)
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Probes:        probes,
		Mutex:         mutex,
	}
	testProbe := &newProbe{}
	probe1 := &Probe{
		ProbeName:  "probe1",
		ProbeState: true,
		Interface:  testProbe,
	}
	fakeController.AddNewProbe(probe1)
	expectedProbeList = append(expectedProbeList, probe1)
	tests := map[string]struct {
		actualProbeList   []*Probe
		expectedProbeList []*Probe
	}{
		"add one probe and check if it is present or not": {actualProbeList: fakeController.Probes, expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

func TestListProbe(t *testing.T) {
	probes := make([]*Probe, 0)
	expectedProbeList := make([]*Probe, 0)
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	mutex := &sync.Mutex{}
	fakeController := &Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Probes:        probes,
		Mutex:         mutex,
	}
	testProbe := &newProbe{}
	probe1 := &Probe{
		Priority:   2,
		ProbeName:  "probe1",
		ProbeState: true,
		Interface:  testProbe,
	}
	probe2 := &Probe{
		Priority:   1,
		ProbeName:  "probe2",
		ProbeState: true,
		Interface:  testProbe,
	}
	fakeController.AddNewProbe(probe1)
	fakeController.AddNewProbe(probe2)
	expectedProbeList = append(expectedProbeList, probe2)
	expectedProbeList = append(expectedProbeList, probe1)
	tests := map[string]struct {
		actualProbeList   []*Probe
		expectedProbeList []*Probe
	}{
		"add some probes and check if they are present or not": {actualProbeList: fakeController.ListProbe(), expectedProbeList: expectedProbeList},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedProbeList, test.actualProbeList)
		})
	}
}

func TestSrart(t *testing.T) {
	var msg1 string
	testProbe := &newProbe{}
	probe1 := &Probe{
		ProbeName:  "probe1",
		ProbeState: true,
		Interface:  testProbe,
	}
	go probe1.Start()
	select {
	case res := <-messageChannel:
		msg1 = res
	case <-time.After(1 * time.Second):
		msg1 = ""
	}

	tests := map[string]struct {
		actualMessage   string
		expectedMessage string
	}{
		"comparing message from start method": {actualMessage: msg1, expectedMessage: message},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedMessage, test.actualMessage)
		})
	}
}

func TestFillDiskDetails(t *testing.T) {
	testProbe := &newProbe{}
	probe1 := &Probe{
		ProbeName:  "probe1",
		ProbeState: true,
		Interface:  testProbe,
	}
	actualDisk := NewDiskInfo()
	expectedDisk := NewDiskInfo()
	probe1.FillDiskDetails(actualDisk)
	expectedDisk.HostName = fakeHostName
	expectedDisk.Uuid = fakeObjectMeta.Name
	expectedDisk.Capacity = fakeCapacity.Storage
	expectedDisk.Model = fakeDetails.Model
	expectedDisk.Serial = fakeDetails.Serial
	expectedDisk.Vendor = fakeDetails.Vendor
	expectedDisk.Path = fakeObj.Path
	tests := map[string]struct {
		actualDisk   DiskInfo
		expectedDisk DiskInfo
	}{
		"comparing message from start method": {actualDisk: *actualDisk, expectedDisk: *expectedDisk},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedDisk, test.actualDisk)
		})
	}
}
