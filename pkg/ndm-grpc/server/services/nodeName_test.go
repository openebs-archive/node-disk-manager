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
	"os"
	"testing"

	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
)

// TestFindNodeName tests FindNodeName service
func TestFindNodeName(t *testing.T) {
	os.Setenv("NODE_NAME", "TEST_NODE")

	var ctx context.Context
	var null *protos.Null

	n := NewNode()
	res, err := n.Name(ctx, null)
	if err != nil {
		t.Error("Error in finding node name")
	}
	if res.NodeName != "TEST_NODE" {
		t.Errorf("Expected node name was %v and got %v", "TEST_NODE", res.NodeName)
	}

}
