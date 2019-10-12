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

	"k8s.io/klog"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/spf13/cobra"
)

// NodeDiskManagerOptions defines a type for the options of NDM
type NodeDiskManagerOptions struct {
	KubeConfig string
}

// AddKubeConfigFlag is used to add a config flag
func AddKubeConfigFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "kubeconfig", "", *value,
		"Path to a kube config. Only required if out-of-cluster.")
}

// NewNodeDiskManager creates a new ndm.
func NewNodeDiskManager() (*cobra.Command, error) {
	// Define the options for NDM
	options := NodeDiskManagerOptions{}

	// Create a new command
	cmd := &cobra.Command{
		Use:   "ndm",
		Short: "ndm controls the Node-Disk-Manager ",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(RunNodeDiskManager(cmd), util.Fatal)
		},
	}

	// add the klog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	//add flag for /metrics endpoint port and endpoint path
	cmd.PersistentFlags().String("port", ":9090", "Port to launch HTTP server.")
	cmd.PersistentFlags().String("metricspath", "/metrics", "Endpointpath to get metrics.")

	flag.CommandLine.Parse([]string{})
	cmd.AddCommand(
		NewCmdBlockDevice(), //Add new command on block device
		NewCmdStart(),       //Add new command to start the ndm controller
	)

	// Define the flags allowed in this command & store each option
	// provided as a flag, into Options
	AddKubeConfigFlag(cmd, &options.KubeConfig)

	return cmd, nil
}

// RunNodeDiskManager starts ndm process
func RunNodeDiskManager(cmd *cobra.Command) error {
	klog.Infof("Starting node disk manager ...")

	return nil
}
