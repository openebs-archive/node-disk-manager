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
	"io/ioutil"
	"os"
	"strings"
)

const (
	BlockSubSystem = "block"
	NVMeSubSystem  = "nvme"
)

// deviceSysPath has a device name and its syspath
type deviceSysPath struct {
	// DeviceName is the name of the device like sda, dm-0, nvme0n1
	DeviceName string

	// SysPath is the path to the device in sysfs.
	// 1. /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4
	// 2. /sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1
	// are examples of syspath for device belonging to block and nvme subsystems
	SysPath string
}

// getParent gets the parent of this device if it has parent
func (s deviceSysPath) getParent() (string, bool) {
	parts := strings.Split(s.SysPath, "/")

	var parentBlockDevice string
	ok := false

	// checking for block subsystem, return the next part after subsystem only
	// if the length is greater. This check is to avoid an index out of range panic.
	for i, part := range parts {
		if part == BlockSubSystem {
			if len(parts)-1 >= i+1 && s.DeviceName != parts[i+1] {
				ok = true
				parentBlockDevice = parts[i+1]
			}
			return parentBlockDevice, ok
		}
	}

	// checking for nvme subsystem, return the 2nd item in hierarchy, which will be the
	// nvme namespace. Length checking is to avoid index out of range in case of malformed
	// links (extremely rare case)
	for i, part := range parts {
		if part == NVMeSubSystem {
			if len(parts)-1 >= i+2 && s.DeviceName != parts[i+2] {
				ok = true
				parentBlockDevice = parts[i+2]
			}
			return parentBlockDevice, ok
		}
	}

	return parentBlockDevice, ok
}

// getPartitions gets the partitions of this device if it has any
func (s deviceSysPath) getPartitions() ([]string, bool) {

	// if partition file has value 0, or the file doesn't exist,
	// can return from there itself
	// partitionPath := s.SysPath + "/partition"
	// if _, err := os.Stat(partitionPath); os.IsNotExist(err) {
	// }

	partitions := make([]string, 0)

	files, err := ioutil.ReadDir(s.SysPath)
	if err != nil {
		return nil, false
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), s.DeviceName) {
			partitions = append(partitions, file.Name())
		}
	}

	return partitions, true
}

// getHolders gets the devices that are held by this device
func (s deviceSysPath) getHolders() ([]string, bool) {
	holderPath := s.SysPath + "/holders"
	holders := make([]string, 0)

	// check if holders are available for this device
	if _, err := os.Stat(holderPath); os.IsNotExist(err) {
		return nil, false
	}

	files, err := ioutil.ReadDir(holderPath)
	if err != nil {
		return nil, false
	}

	for _, file := range files {
		holders = append(holders, file.Name())
	}
	return holders, true
}

// getSlaves gets the devices to which this device is a slave. Or, the devices
// which holds this device
func (s deviceSysPath) getSlaves() ([]string, bool) {
	slavePath := s.SysPath + "/slaves"
	slaves := make([]string, 0)

	// check if slaves are available for this device
	if _, err := os.Stat(slavePath); os.IsNotExist(err) {
		return nil, false
	}

	files, err := ioutil.ReadDir(slavePath)
	if err != nil {
		return nil, false
	}

	for _, file := range files {
		slaves = append(slaves, file.Name())
	}
	return slaves, true
}
