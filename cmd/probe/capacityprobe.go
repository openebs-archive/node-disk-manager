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

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/controller"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	capacityProbePriority = 3
	capacityConfigKey     = "capacity-probe"
)

var (
	capacityProbeName  = "capacity probe"
	capacityProbeState = defaultEnabled
)

var capacityProbeRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		glog.Error("unable to configure", capacityProbeName)
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

// FillDiskDetails updates the capacity of the disk , if the capacity is
// not populated.
func (cp *capacityProbe) FillDiskDetails(d *controller.DiskInfo) {

	if d.Capacity != 0 {
		return
	}
	sysPath := d.ProbeIdentifiers.UdevIdentifier
	blockSize, err := strconv.ParseInt(d.NoOfBlocks, 10, 64)
	if err != nil {
		glog.Error("unable to parse the block size ", err)
		return
	}
	b, err := ioutil.ReadFile(sysPath + "/queue/hw_sector_size")
	if err != nil {
		glog.Error("unable to read sector size", err)
		return
	}
	sectorSize, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		glog.Error("unable to parse the sector size", err)
		return
	}
	d.Capacity = uint64(blockSize * sectorSize)
}
