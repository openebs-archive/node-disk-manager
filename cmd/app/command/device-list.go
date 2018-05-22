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
	"github.com/spf13/cobra"
)

// NewSubCmdListBlockDevice is to list block device is created
func NewSubCmdListBlockDevice() *cobra.Command {
	options := CmdStartOptions{}
	getCmd := &cobra.Command{
		Use:   "list",
		Short: "List block devices",
		Long: `the set of block devices on the node
		can be listed via 'ndm device list' command`,
		Run: func(cmd *cobra.Command, args []string) {
			controller.DeviceList(options.kubeconfig)
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
