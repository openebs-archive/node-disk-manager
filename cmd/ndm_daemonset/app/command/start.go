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
	"fmt"
	"os"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/filter"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/probe"
	"github.com/spf13/cobra"
)

//CmdStartOptions options for start command
type CmdStartOptions struct {
	kubeconfig string
}

//NewCmdStart starts the ndm controller
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "Node disk controller",
		Long:  ` watches for ndm custom resources via "ndm start" command `,
		Run: func(cmd *cobra.Command, args []string) {
			ctrl, err := controller.NewController(options.kubeconfig)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// Broadcast starts broadcasting controller pointer. Using this
			// each probe and filter registers themselves.
			ctrl.Broadcast()
			// Start starts registering of filters present in RegisteredFilters
			filter.Start(filter.RegisteredFilters)
			// Start starts registering of probes present in RegisteredProbes
			probe.Start(probe.RegisteredProbes)
			ctrl.Start()
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	getCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})

	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)
	return getCmd
}
