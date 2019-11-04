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

var sysFSDirectoryPath = "/sys/"

// Device represents a blockdevice. This struct is used by hierachy pkg which is used to
// get the necessary blockdevice hierarchy information
type Device struct {
	// Path of the blockdevice. eg: /dev/sda, /dev/dm-0
	Path string
}

// DependentDevices represents all the dependent blockdevices of the given Device
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

// GetDependents gets all the dependent devices for a given Device
func (d *Device) GetDependents() (DependentDevices, error) {
	dependents := DependentDevices{}

	// get the syspath of the blockdevice
	blockDeviceSysPath, err := getDeviceSysPath(d.Path)
	if err != nil {
		return dependents, err
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

	// adding /devprefix to partition, slaves and holders
	dependents.Partitions = addDevPrefix(dependents.Partitions)
	dependents.Holders = addDevPrefix(dependents.Holders)
	dependents.Slaves = addDevPrefix(dependents.Slaves)

	return dependents, nil
}

// addDevPrefix adds the /dev prefix to all the device names
func addDevPrefix(paths []string) []string {
	result := make([]string, 0)
	for i := range paths {
		result = append(result, "/dev/"+paths[i])
	}
	return result
}
