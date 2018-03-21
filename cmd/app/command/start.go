/*
Copyright 2018 The OpenEBS Authors.

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

package command

import (
	goflag "flag"

	"github.com/openebs/node-disk-manager/cmd/controller"
	"github.com/openebs/node-disk-manager/cmd/metrics"
	"github.com/spf13/cobra"
)

//NewCmdStart starts the ndm controller
func NewCmdStart() *cobra.Command {
	//var target string
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "Node disk controller",
		Long:  ` watches for ndm custom resources via "ndmctl start" command `,
		Run: func(cmd *cobra.Command, args []string) {
			// Start HTTP server for /metrics endpoint 
			go metrics.StartHttpServer()
			controller.Watch()
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	getCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})

	return getCmd
}
