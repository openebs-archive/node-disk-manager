/*
Copyright 2018 OpenEBS Authors.

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
	capacityProbePriority = 4
	capacityConfigKey     = "capacity-probe"
	// sectorSize is the sector size as understood by the unix systems. It is kept as 512 bytes.
	// all entries in /sys/class/block/sda/size are in 512 byte blocks
	sectorSize = 512
)

var (
	capacityProbeName  = "capacity probe"
	capacityProbeState = defaultEnabled
)

var capacityProbeRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", capacityProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == capacityConfigKey {
				capacityProbeName = probeConfig.Name
				capacityProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}

	newRegistryProbe := &registerProbe{
		priority:   capacityProbePriority,
		name:       capacityProbeName,
		state:      capacityProbeState,
		pi:         newCapacityProbe(),
		controller: ctrl,
	}
	newRegistryProbe.register()
}

// capacityProbe fills the capacity of the disk
type capacityProbe struct {
}

func newCapacityProbe() *capacityProbe {
	return &capacityProbe{}
}

// It is part of probe interface. Hence, empty implementation.
func (cp *capacityProbe) Start() {}

// FillBlockDeviceDetails updates the capacity of the disk , if the capacity is
// not populated.
func (cp *capacityProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {

	if blockDevice.Capacity.Storage != 0 {
		return
	}
	sysPath := blockDevice.SysPath
	b, err := ioutil.ReadFile(sysPath + "/size")
	if err != nil {
		klog.Error("unable to parse the block size ", err)
		return
	}
	blockSize, err := strconv.ParseInt(strings.TrimSuffix(string(b), "\n"), 10, 64)
	if err != nil {
		klog.Error("unable to parse the block size", err)
		return
	}

	// The size (/size) entry returns the `nr_sects` field of the block device structure.
	// Ref: https://elixir.bootlin.com/linux/v4.4/source/fs/block_dev.c#L1267
	//
	// Traditionally, in Unix disk size contexts, “sector” or “block” means 512 bytes,
	// regardless of what the manufacturer of the underlying hardware might call a “sector” or “block”
	// Ref: https://elixir.bootlin.com/linux/v4.4/source/fs/block_dev.c#L487
	//
	// Therefore, to get the capacity of the device it needs to always multiplied with 512
	blockDevice.Capacity.Storage = uint64(blockSize * sectorSize)

	if blockDevice.DeviceAttributes.LogicalBlockSize == 0 {
		blockDevice.DeviceAttributes.LogicalBlockSize = uint32(sectorSize)
	}
}
