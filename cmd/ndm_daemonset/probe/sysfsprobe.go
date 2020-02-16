/*
Copyright 2020 OpenEBS Authors.

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

/*

Find queue sysfs doc here.
https://www.kernel.org/doc/Documentation/block/queue-sysfs.txt

*/

package probe

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

const (
	sysfsProbePriority = 4
	sysfsConfigKey     = "sysfs-probe"
)

var (
	sysfsProbeName  = "sysfs probe"
	sysfsProbeState = defaultEnabled
)

var sysfsProbeRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", sysfsProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == sysfsConfigKey {
				sysfsProbeName = probeConfig.Name
				sysfsProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}

	newRegistryProbe := &registerProbe{
		priority:   sysfsProbePriority,
		name:       sysfsProbeName,
		state:      sysfsProbeState,
		pi:         newSysFSProbe(),
		controller: ctrl,
	}
	newRegistryProbe.register()
}

// sysfsProbe fills the logical sector size,
// physical sector size, drive type(ssd or hdd) of the disk
type sysfsProbe struct {
}

func newSysFSProbe() *sysfsProbe {
	return &sysfsProbe{}
}

// It is part of probe interface. Hence, empty implementation.
func (cp *sysfsProbe) Start() {}

// FillBlockDeviceDetails updates the logical sector size,
// physical sector size, drive type(ssd or hdd) of the disk
// if those are not populated.
func (cp *sysfsProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	sysPath := blockDevice.SysPath

	if blockDevice.DeviceAttributes.LogicalBlockSize == 0 {
		logicalBlockSize, err := readSysFSFileAsInt64(sysPath + "/queue/logical_block_size")
		if err != nil {
			klog.Warning("unable to read logical block size", err)
		} else if logicalBlockSize != 0 {
			blockDevice.DeviceAttributes.LogicalBlockSize = uint32(logicalBlockSize)
			klog.Infof("blockdevice path: %s logical sector size :%d filled by queuesysfs probe.",
				blockDevice.DevPath, blockDevice.DeviceAttributes.LogicalBlockSize)
		}
	}

	if blockDevice.DeviceAttributes.PhysicalBlockSize == 0 {
		physicalBlockSize, err := readSysFSFileAsInt64(sysPath + "/queue/physical_block_size")
		if err != nil {
			klog.Warning("unable to read physical block size", err)
		} else if physicalBlockSize != 0 {
			blockDevice.DeviceAttributes.PhysicalBlockSize = uint32(physicalBlockSize)
			klog.Infof("blockdevice path: %s physical sector size :%d filled by sysfs probe.",
				blockDevice.DevPath, blockDevice.DeviceAttributes.PhysicalBlockSize)
		}
	}

	if blockDevice.DeviceAttributes.DriveType == "" {
		rotational, err := readSysFSFileAsInt64(sysPath + "/queue/rotational")
		if err != nil {
			klog.Warning("unable to read rotational value", err)
		} else {
			if rotational == 1 {
				blockDevice.DeviceAttributes.DriveType = "HDD"

			} else if rotational == 0 {
				blockDevice.DeviceAttributes.DriveType = "SSD"
			}
			klog.Infof("blockdevice path: %s drive type :%s filled by sysfs probe.",
				blockDevice.DevPath, blockDevice.DeviceAttributes.DriveType)
		}
	}

	if blockDevice.Capacity.Storage == 0 {
		hwSectorSize, err := readSysFSFileAsInt64(sysPath + "/queue/hw_sector_size")
		if err != nil {
			klog.Warning("unable to read hw sector size", err)
			return
		}
		blockSize, err := readSysFSFileAsInt64(sysPath + "/size")
		if err != nil {
			klog.Warning("unable to read block size", err)
			return
		} else if blockSize != 0 {
			blockDevice.Capacity.Storage = uint64(uint32(blockSize)) * uint64(uint32(hwSectorSize))
			klog.Infof("blockdevice path: %s capacity :%d filled by sysfs probe.",
				blockDevice.DevPath, blockDevice.Capacity.Storage)
		}
	}
}

// readSysFSFileAsInt64 reads a file and
// converts that content into int64
func readSysFSFileAsInt64(sysPath string) (int64, error) {
	b, err := ioutil.ReadFile(sysPath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSuffix(string(b), "\n"), 10, 64)
}
