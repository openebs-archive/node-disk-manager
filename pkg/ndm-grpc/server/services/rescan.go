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

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/probe"
	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// Rescan sync etcd with ndm
func (n *Node) Rescan(ctx context.Context, null *protos.Null) (*protos.Message, error) {

	klog.Info("Rescan initiated")

	ctrl, err := controller.NewController()
	if err != nil {
		klog.Errorf("Error creating a controller %v", err)
		return nil, status.Errorf(codes.NotFound, "Namespace not found")
	}

	err = ctrl.SetControllerOptions(controller.NDMOptions{ConfigFilePath: "/host/node-disk-manager.config"})
	if err != nil {
		klog.Errorf("Error setting config to controller %v", err)
		return nil, status.Errorf(codes.Internal, "Error setting config to controller")
	}

	err = probe.Rescan(ctrl)
	if err != nil {
		klog.Errorf("Rescan failed %v", err)
		return nil, status.Errorf(codes.AlreadyExists, "Rescan failed")
	}

	return &protos.Message{Msg: "Rescan initiated, check the logs for more info"}, nil
}
