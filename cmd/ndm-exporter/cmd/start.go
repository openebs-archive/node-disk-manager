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
	"github.com/openebs/node-disk-manager/ndm-exporter"
	"github.com/openebs/node-disk-manager/pkg/util"
	"github.com/spf13/cobra"
)

var exporter ndm_exporter.Exporter

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the exporter",
	Run: func(cmd *cobra.Command, args []string) {
		util.CheckErr(exporter.Run(), util.Fatal)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().StringVar(&exporter.Mode, "mode",
		ndm_exporter.ClusterLevel,
		`Mode in which the exporter need to be started (cluster / node)`)

	startCmd.PersistentFlags().StringVar(&exporter.Server.ListenPort, "port",
		ndm_exporter.Port,
		"Port on which metrics is available")

	startCmd.PersistentFlags().StringVar(&exporter.Server.MetricsPath, "metrics",
		ndm_exporter.MetricsPath,
		"The URL end point at which metrics is available (/metrics, /endpoint)")
}
