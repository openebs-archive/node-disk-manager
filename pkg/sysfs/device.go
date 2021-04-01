/*
Copyright 2020 The OpenEBS Authors

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

package sysfs

import (
	"fmt"
	"strings"
)

// Device represents a blockdevice using its sysfs path.
type Device struct {
	// deviceName is the name of the device node sda, sdb, dm-0 etc
	deviceName string

	// Path of the blockdevice. eg: /dev/sda, /dev/dm-0
	path string

	// SysPath of the blockdevice. eg: /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/
	sysPath string
}

// NewSysFsDeviceFromDevPath is used to get sysfs device struct from the device devpath
// The sysfs device struct contains the device name along with the syspath
func NewSysFsDeviceFromDevPath(devPath string) (*Device, error) {
	devName := strings.Replace(devPath, "/dev/", "", 1)
	if len(devName) == 0 {
		return nil, fmt.Errorf("unable to create sysfs device from devPath for device: %s, error: device name empty")
	}

	sysPath, err := getDeviceSysPath(devPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create sysfs device from devpath for device: %s, error: %v", devPath, err)
	}

	dev := &Device{
		deviceName: devName,
		path:       devPath,
		sysPath:    sysPath,
	}
	return dev, nil
}
