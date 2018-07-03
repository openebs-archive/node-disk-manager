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
	"time"

	"github.com/openebs/node-disk-manager/cmd/controller"
	ndmFakeClientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

var messageChannel = make(chan string)
var message = "This is a message from start method"

type probe1 struct {
	name string
}

func (p1 *probe1) Start() {
	messageChannel <- message
}

func (p1 *probe1) FillDiskDetails(*controller.DiskInfo) {
	messageChannel <- message
}

func TestRegisterProbe(t *testing.T) {
	probes := make([]*controller.Probe, 0)
	fakeHostName := "fake-host-name"
	fakeNdmClient := ndmFakeClientset.NewSimpleClientset()
	fakeKubeClient := fake.NewSimpleClientset()
	mutex := &sync.Mutex{}
	fakeController := &controller.Controller{
		HostName:      fakeHostName,
		KubeClientset: fakeKubeClient,
		Clientset:     fakeNdmClient,
		Probes:        probes,
		Mutex:         mutex,
	}

	var msg1, msg2 string
	var i controller.ProbeInterface = &probe1{name: "probe1"}
	newPrgisterProbe1 := &registerProbe{
		priority:       1,
		probeName:      "probe-1",
		probeState:     true,
		probeInterface: i,
		controller:     fakeController,
	}
	go newPrgisterProbe1.register()
	select {
	case res := <-messageChannel:
		msg1 = res
	case <-time.After(1 * time.Second):
		msg1 = ""
	}

	newPrgisterProbe2 := &registerProbe{
		priority:       1,
		probeName:      "probe-2",
		probeState:     false,
		probeInterface: i,
		controller:     fakeController,
	}
	go newPrgisterProbe2.register()
	select {
	case res := <-messageChannel:
		msg2 = res
	case <-time.After(2 * time.Second):
		msg2 = ""
	}

	tests := map[string]struct {
		actualMessage   string
		expectedMessage string
	}{
		"probe status is enabled so it receives actual message": {actualMessage: msg1, expectedMessage: message},
		"probe status is disabled so it receives empty message": {actualMessage: msg2, expectedMessage: ""},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedMessage, test.actualMessage)
		})
	}
}
