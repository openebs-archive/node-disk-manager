/*
Copyright 2019 The OpenEBS Authors
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

package services

import (
	"context"
	"testing"

	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
	"k8s.io/klog"
)

// TestGetParentDisks tests the GetParentDisks function
func TestListBlockDeviceDetails(t *testing.T) {

	n := NewNode()

	mockDevice := &protos.BlockDevice{
		Name: "/dev/sda",
		Type: "Disk",
	}

	var ctx context.Context
	diskinfo, err := n.ListBlockDeviceDetails(ctx, mockDevice)
	if err != nil {
		t.Errorf("Error listing details %v", err)
	}
	klog.Info(diskinfo)

}
