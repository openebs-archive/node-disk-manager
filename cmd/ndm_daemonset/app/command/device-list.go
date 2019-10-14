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
	"html/template"
	"os"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/spf13/cobra"
)

/*
defaultDeviceList template use to print list of device
This template looks like below -

when disk resources present
root@instance-1:~#ndm device list
NAME                                         PATH      CAPACITY       STATUS    SERIAL                   MODEL               VENDOR
disk-ccc636c88bd9ab09dde9de476309058d        /dev/sda  10737418240    Inactive  instance-1               PersistentDisk      Google
disk-ce41f8f5fa22acb79ec56292441dc207        /dev/sdb  10737418240    Active    disk-1                   PersistentDisk      Google

when no resource present
root@instance-1:~#ndm device list
No disk resource present.
*/
const defaultDeviceList = `
{{- if .Items}}
	{{- printf "%-45s" "NAME"}}
	{{- printf "%-10s" "PATH"}}
	{{- printf "%-15s" "CAPACITY"}}
	{{- printf "%-10s" "STATUS"}}
	{{- printf "%-25s" "SERIAL"}}
	{{- printf "%-20s" "MODEL"}}
	{{- printf "%-20s" "VENDOR"}}
{{range .Items}}
	{{- printf "%-45s" .ObjectMeta.Name}}
	{{- printf "%-10s" .Spec.Path}}
	{{- printf "%-15d" .Spec.Capacity.Storage}}
	{{- printf "%-10s" .Status.State}}
	{{- printf "%-25s" .Spec.Details.Serial}}
	{{- printf "%-20s" .Spec.Details.Model}}
	{{- printf "%-20s" .Spec.Details.Vendor}}
{{end}}
{{- else}}
	{{- printf "%s" "No disk resource present."}}
{{end}}`

// NewSubCmdListBlockDevice is to list block device is created
func NewSubCmdListBlockDevice() *cobra.Command {
	options := CmdStartOptions{}
	getCmd := &cobra.Command{
		Use:   "list",
		Short: "List block devices",
		Long: `the set of block devices on the node
		can be listed via 'ndm device list' command`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deviceList(options.kubeconfig)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	goflag.CommandLine.Parse([]string{})

	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)

	return getCmd
}

// deviceList prints list of devices using defaultDeviceList template
func deviceList(kubeconfig string) error {
	ctrl, err := controller.NewController(kubeconfig)
	if err != nil {
		return err
	}
	diskList, err := ctrl.ListDiskResource()
	if err != nil {
		return err
	}
	diskListTemplate := template.Must(template.New("defaultDeviceList").Parse(defaultDeviceList))
	err = diskListTemplate.Execute(os.Stdout, diskList)
	if err != nil {
		return err
	}
	return nil
}
