/*
Copyright 2020 The OpenEBS Authors.

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

package grpc

import (
	"net"
	"os"

	protos "github.com/openebs/node-disk-manager/pkg/ndm-grpc/protos/ndm"
	"github.com/openebs/node-disk-manager/pkg/ndm-grpc/server/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog"
)

//Start starts the grpc server
func Start() {
	{

		// Creating a grpc server, use WithInsecure to allow http connections
		gs := grpc.NewServer()

		// Creates an instance of Info
		is := services.NewInfo()

		// Creates an instance of Service
		ss := services.NewService()

		// Creates an instance of Node
		ns := services.NewNode()

		// This helps clients determine which services are available to call
		reflection.Register(gs)

		// Similar to registring handlers for http
		protos.RegisterInfoServer(gs, is)

		protos.RegisterISCSIServer(gs, ss)

		protos.RegisterNodeServer(gs, ns)

		l, err := net.Listen("tcp", "0.0.0.0:9090")
		if err != nil {
			klog.Errorf("Unable to listen %f", err)
			os.Exit(1)
		}

		// Listen for requests
		klog.Info("Starting server at 9090")
		gs.Serve(l)

	}

}
