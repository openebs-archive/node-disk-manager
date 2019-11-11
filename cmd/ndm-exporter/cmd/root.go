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

package cmd

import (
	goflag "flag"
	"fmt"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/openebs/node-disk-manager/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "exporter",
	Short: "exporter can be used to expose block device metrics",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		util.CheckErr(RunNodeDiskExporter(cmd), util.Fatal)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// RunNodeDiskExporter logs the starting of NDM exporter
func RunNodeDiskExporter(cmd *cobra.Command) error {
	klog.Infof("Starting NDM Exporter...")
	klog.Infof("Version Tag : %v", version.GetVersion())
	klog.Infof("GitCommit : %v", version.GetGitCommit())
	return nil
}

func init() {
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	// HACK: without the following line, the logs will be prefixed with an error
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	_ = goflag.CommandLine.Parse([]string{})
}
