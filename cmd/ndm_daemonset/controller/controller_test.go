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
	set environment variable "NODE_NAME" with some value and getNodeName
	unset environment variable "NODE_NAME" with some value and getNodeName
*/
func TestGetNodeName(t *testing.T) {
	fakeNodeName := "fake-node-name"
	nodeName1, err1 := getNodeName()
	os.Setenv("NODE_NAME", fakeNodeName)
	nodeName2, err2 := getNodeName()
	expectedErr2 := errors.New("error getting node name")
	tests := map[string]struct {
		actualNodeName   string
		expectedNodeName string
		actualError      error
		expectedError    error
	}{
		"call getNodeName when env variable not present": {actualNodeName: nodeName1, actualError: err1, expectedNodeName: "", expectedError: expectedErr2},
		"call getNodeName when env variable present":     {actualNodeName: nodeName2, actualError: err2, expectedNodeName: fakeNodeName, expectedError: nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedNodeName, test.actualNodeName)
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

func TestController_setNodeLabels(t *testing.T) {
	fakeNdmClient := CreateFakeClient(t)
	nodeAttributes := make(map[string]string, 0)
	nodeAttributes[NodeNameKey] = fakeHostName
	fakeController := &Controller{
		NodeAttributes: nodeAttributes,
		Clientset:      fakeNdmClient,
	}
	os.Setenv("LABEL_LIST", "fake-label,node-attributes")
	tests := map[string]struct {
		ExpectedController *Controller
		wantErr            bool
	}{
		"1": {
			ExpectedController: &Controller{
				NodeAttributes: map[string]string{
					NodeNameKey:              fakeHostName,
					"kubernetes.io/hostname": fakeHostName,
					"kubernetes.io/arch":     "fake-arch",
					HostNameKey:              fakeHostName,
				},
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := fakeController.setNodeLabels(); (err != nil) != tt.wantErr {
				t.Errorf("setNodeLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.ExpectedController.NodeAttributes, fakeController.NodeAttributes)
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
