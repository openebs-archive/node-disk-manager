/*
Copyright 2020 The OpenEBS Authors
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
	"strings"

	protos "github.com/openebs/node-disk-manager/spec/ndm"

	ps "github.com/mitchellh/go-ps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

const iscsiServiceName = "iscsid"

// ISCSIStatus gives the status of iSCSI service
func (n *Node) ISCSIStatus(ctx context.Context, null *protos.Null) (*protos.Status, error) {

	klog.Info("Finding ISCSI status")

	// This will fetch the processes regardless of which OS is being used
	processList, err := ps.Processes()
	if err != nil {
		klog.Error(err)
		return nil, status.Errorf(codes.Internal, "Error fetching the processes")
	}

	var found bool

	for _, p := range processList {

		if strings.Contains(p.Executable(), iscsiServiceName) {
			klog.Infof("%v is running with process id %v", p.Executable(), p.Pid())
			found = true
		}
	}

	if !found {
		// Note: When using clients like grpcurl, they might return empty output as response when converting to json
		// Set the appropriate flags to avoid that. In case of grpcurl, it is -emit-defaults
		return &protos.Status{Status: false}, nil
	}

	return &protos.Status{Status: true}, nil

}
