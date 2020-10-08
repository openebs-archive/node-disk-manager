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
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/sysfs"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

const (
	sysfsProbePriority = 2
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

	sysFsDevice, err := sysfs.NewSysFsDeviceFromDevPath(blockDevice.DevPath)
	if err != nil {
		klog.Errorf("unable to get sysfs device for device: %s, err: %v", blockDevice.DevPath, err)
		return
	}

	if blockDevice.DeviceAttributes.LogicalBlockSize == 0 {
		logicalBlockSize, err := sysFsDevice.GetLogicalBlockSize()
		if err != nil {
			klog.Warningf("unable to get logical block size for device: %s, err: %v", blockDevice.DevPath, err)
		} else if logicalBlockSize == 0 {
			klog.Warningf("logical block size for device: %s reported as 0", blockDevice.DevPath)
		}
		blockDevice.DeviceAttributes.LogicalBlockSize = uint32(logicalBlockSize)
		klog.V(4).Infof("blockdevice path: %s logical block size :%d filled by sysfs probe.",
			blockDevice.DevPath, blockDevice.DeviceAttributes.LogicalBlockSize)
	}

	if blockDevice.DeviceAttributes.PhysicalBlockSize == 0 {
		physicalBlockSize, err := sysFsDevice.GetPhysicalBlockSize()
		if err != nil {
			klog.Warningf("unable to get physical block size for device: %s, err: %v", blockDevice.DevPath, err)
		} else if physicalBlockSize == 0 {
			klog.Warningf("physical block size for device: %s reported as 0", blockDevice.DevPath)
		}
		blockDevice.DeviceAttributes.PhysicalBlockSize = uint32(physicalBlockSize)
		klog.V(4).Infof("blockdevice path: %s physical block size :%d filled by sysfs probe.",
			blockDevice.DevPath, blockDevice.DeviceAttributes.PhysicalBlockSize)
	}

	if blockDevice.DeviceAttributes.HardwareSectorSize == 0 {
		hwSectorSize, err := sysFsDevice.GetPhysicalBlockSize()
		if err != nil {
			klog.Warningf("unable to get hardware sector size for device: %s, err: %v", blockDevice.DevPath, err)
		} else if hwSectorSize == 0 {
			klog.Warningf("hardware sector size for device: %s reported as 0", blockDevice.DevPath)
		}
		blockDevice.DeviceAttributes.HardwareSectorSize = uint32(hwSectorSize)
		klog.V(4).Infof("blockdevice path: %s hardware sector size :%d filled by sysfs probe.",
			blockDevice.DevPath, blockDevice.DeviceAttributes.HardwareSectorSize)
	}

	if blockDevice.DeviceAttributes.DriveType == "" {
		driveType, err := sysFsDevice.GetDriveType()
		if err != nil {
			klog.Warningf("unable to get drive type for device: %s, err: %v", blockDevice.DevPath, err)
		}
		blockDevice.DeviceAttributes.DriveType = driveType
		klog.V(4).Infof("blockdevice path: %s drive type :%s filled by sysfs probe.",
			blockDevice.DevPath, blockDevice.DeviceAttributes.DriveType)
	}

	if blockDevice.Capacity.Storage == 0 {
		capacity, err := sysFsDevice.GetCapacityInBytes()
		if err != nil {
			klog.Warningf("unable to get capacity for device: %s, err: %v", blockDevice.DevPath, err)
		}
		blockDevice.Capacity.Storage = uint64(capacity)
		klog.V(4).Infof("blockdevice path: %s capacity :%d filled by sysfs probe.",
			blockDevice.DevPath, blockDevice.Capacity.Storage)
	}
}
