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

package hierarchy

import (
	"path/filepath"
	"strings"
)

var sysFSDirectoryPath = "/sys/"

type Device struct {
	// Path of the blockdevice. eg: /dev/sda, /dev/dm-0
	Path string
}

type DependentDevices struct {
	// Parent is the parent device of the given blockdevice
	Parent string

	// Partitions are the partitions of this device if any
	Partitions []string

	// Holders is the slice of block-devices that are held by a given
	// blockdevice
	Holders []string

	// Slaves is the slice of blockdevices to which the given blockdevice
	// is a slave
	Slaves []string
}

func (d *Device) GetDependents() (DependentDevices, error) {
	dependents := DependentDevices{}
	blockDeviceName := strings.Replace(d.Path, "/dev/", "", 1)

	blockDeviceSymLink := sysFSDirectoryPath + "class/block/" + blockDeviceName

	sysPath, err := filepath.EvalSymlinks(blockDeviceSymLink)
	if err != nil {
		return dependents, err
	}
	blockDeviceSysPath := deviceSysPath{
		SysPath:    sysPath,
		DeviceName: blockDeviceName,
	}

	// parent device
	if parent, ok := blockDeviceSysPath.getParent(); ok {
		dependents.Parent = parent
	}

	// get the partitions
	if partitions, ok := blockDeviceSysPath.getPartitions(); ok {
		dependents.Partitions = partitions
	}

	// get the holder devices
	if holders, ok := blockDeviceSysPath.getHolders(); ok {
		dependents.Holders = append(dependents.Holders, holders...)
	}

	// get the slaves
	if slaves, ok := blockDeviceSysPath.getSlaves(); ok {
		dependents.Slaves = append(dependents.Slaves, slaves...)
	}

	// adding /dev prefix
	dependents.Parent = "/dev/" + dependents.Parent

	// adding /dev prefix
	for i, _ := range dependents.Partitions {
		dependents.Partitions[i] = "/dev/" + dependents.Partitions[i]
	}

	// adding /dev prefix
	for i, _ := range dependents.Slaves {
		dependents.Slaves[i] = "/dev/" + dependents.Slaves[i]
	}

	// adding /dev prefix
	for i, _ := range dependents.Holders {
		dependents.Holders[i] = "/dev/" + dependents.Holders[i]
	}

	return dependents, nil
}
