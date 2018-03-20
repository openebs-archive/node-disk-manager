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
	"flag"
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/spf13/cobra"
)

// NewNodeDiskManager creates a new ndmctl.
func NewNodeDiskManager() (*cobra.Command, error) {
	// Create a new command
	cmd := &cobra.Command{
		Use:   "ndmctl",
		Short: "ndmctl controls the Node-Disk-Manager ",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(RunNodeDiskManager(cmd), util.Fatal)
		},
	}
	// add the glog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	flag.CommandLine.Parse([]string{})
	cmd.AddCommand(
		NewCmdBlockDevice(), //Add new command on block device
		NewCmdStart(), //Add new command to start the ndm controller
	)

	return cmd, nil
}

// Run ndmctl
func RunNodeDiskManager(cmd *cobra.Command) error {
	glog.Infof("Starting node disk manager ...")

	return nil
}
