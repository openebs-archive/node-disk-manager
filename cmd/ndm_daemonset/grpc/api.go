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

	"github.com/openebs/node-disk-manager/api-service/node/services"
	protos "github.com/openebs/node-disk-manager/spec/ndm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog"
)

const (
	// IP represents the IP address of the api service
	IP = "0.0.0.0"
	// Port represents the port of the API service
	Port = "9115"
	// DefaultAddress represents the DefaultAddress at which api service can be called on which is 0.0.0.0:9115
	DefaultAddress = IP + ":" + Port
)

// Address represents the address given by user
var Address = ""

// Start starts the grpc server
func Start() {
	{
		// Creating a grpc server, use WithInsecure to allow http connections
		grpcServer := grpc.NewServer()

		// Creates an instance of Info
		infoService := services.NewInfo()

		// Creates an instance of Node
		nodeService := services.NewNode()

		// This helps clients determine which services are available to call
		reflection.Register(grpcServer)

		// Similar to registering handlers for http
		protos.RegisterInfoServer(grpcServer, infoService)

		protos.RegisterNodeServer(grpcServer, nodeService)

		l, err := net.Listen("tcp", Address)
		if err != nil {
			klog.Errorf("Unable to listen %v", err)
			os.Exit(1)
		}

		// Listen for requests
		klog.Infof("Starting server at : %v ", Address)
		err = grpcServer.Serve(l)
		if err != nil {
			klog.Errorf("Unable to Serve %v", err)
			os.Exit(1)
		}

	}
}
