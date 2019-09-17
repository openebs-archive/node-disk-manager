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
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/spf13/cobra"
)

var exporter Exporter

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		glog.Info("Mode:", mode)
		if mode != ClusterLevel && mode != NodeLevel {
			cmd.Printf("unknown mode %s selected for starting exporter", mode)
			return
		}
		exporter.Mode = mode
		exporter.Port, _ = cmd.Flags().GetString("port")
		exporter.MetricsPath, _ = cmd.Flags().GetString("metrics")
		util.CheckErr(exporter.Run(), util.Fatal)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().StringVar(&exporter.Mode, "mode",
		ClusterLevel,
		`Mode in which the exporter need to be started (cluster / node)`)

	startCmd.PersistentFlags().StringVar(&exporter.Port, "port",
		Port,
		"Port on which metrics is available")

	startCmd.PersistentFlags().StringVar(&exporter.MetricsPath, "metrics",
		MetricsPath,
		"The URL end point at which metrics is available (/metrics, /endpoint)")
}