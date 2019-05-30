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
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
	set environment variable "NODE_NAME" with some value and setNodeName
	unset environment variable "NODE_NAME" with some value and setNodeName
*/
func TestSetNodeName(t *testing.T) {
	fakeNodeName := "fake-node-name"
	ctrl1 := &Controller{}
	ctrl2 := &Controller{}
	err1 := ctrl1.setNodeName()
	os.Setenv("NODE_NAME", fakeNodeName)
	err2 := ctrl2.setNodeName()
	expectedErr2 := errors.New("error building hostname")
	tests := map[string]struct {
		actualController *Controller
		expectedHostName string
		actualError      error
		expectedError    error
	}{
		"call setNodeName when env variable not present": {actualController: ctrl1, actualError: err1, expectedHostName: "", expectedError: expectedErr2},
		"call setNodeName when env variable present":     {actualController: ctrl2, actualError: err2, expectedHostName: fakeNodeName, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedHostName, test.actualController.HostName)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

func TestSetNamespace(t *testing.T) {
	fakeNamespace := "openebs"
	ns1, err1 := getNamespace()
	os.Setenv("NAMESPACE", fakeNamespace)
	ns2, err2 := getNamespace()
	expectedErr2 := errors.New("error getting namespace")
	tests := map[string]struct {
		actualNamespace   string
		expectedNamespace string
		actualError       error
		expectedError     error
	}{
		"call setNamespace when env variable not present": {actualNamespace: ns1, actualError: err1, expectedNamespace: "", expectedError: expectedErr2},
		"call setNamespace when env variable present":     {actualNamespace: ns2, actualError: err2, expectedNamespace: fakeNamespace, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedNamespace, test.actualNamespace)
			assert.Equal(t, test.expectedError, test.actualError)
		})
	}
}

/*
	Broadcast start broadcasting controller pointer in ControllerBroadcastChannel channel
	In this test case read ControllerBroadcastChannel channel and match controller pointer
*/
func TestBroadcast(t *testing.T) {
	ctrl := &Controller{}
	ctrl.Broadcast()
	actualController := <-ControllerBroadcastChannel
	tests := map[string]struct {
		actualController   *Controller
		expectedController *Controller
	}{
		"match controller from broadcast channel": {actualController: actualController, expectedController: ctrl},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedController, test.actualController)
		})
	}
}
