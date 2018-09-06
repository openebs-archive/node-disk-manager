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

package udevevent

import (
	"testing"

	"github.com/openebs/node-disk-manager/cmd/controller"
)

func TestNewMonitor(t *testing.T) {
	fakeController := &controller.Controller{}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	monitor, err := newMonitor()
	if err != nil {
		t.Error(err)
	}
	defer monitor.free()
	if monitor.udev == nil {
		t.Errorf("udev should not be nil")
	}
	if monitor.udevMonitor == nil {
		t.Errorf("udevMonitor should not be nil")
	}
}

func TestSetup(t *testing.T) {
	fakeController := &controller.Controller{}
	go func() {
		controller.ControllerBroadcastChannel <- fakeController
	}()
	monitor, err := newMonitor()
	if err != nil {
		t.Error(err)
	}
	defer monitor.free()
	fd, err := monitor.setup()
	if err != nil {
		t.Error(err)
	}
	if fd < 3 {
		t.Errorf("fd value should be greater than 2")
	}
}
