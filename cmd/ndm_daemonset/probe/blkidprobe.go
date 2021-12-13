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

package probe

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/blkid"
	"k8s.io/klog"
)

const (
	blkidProbePriority = 4
)

var (
	blkidProbeState = defaultEnabled
)

type blkidProbe struct {
}

var blkidProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure custom tag probe")
		return
	}
	probe := &blkidProbe{}

	newRegisterProbe := &registerProbe{
		priority:   blkidProbePriority,
		name:       "blkid probe",
		state:      blkidProbeState,
		pi:         probe,
		controller: ctrl,
	}

	newRegisterProbe.register()
}

func (bp *blkidProbe) Start() {}

func (bp *blkidProbe) FillBlockDeviceDetails(bd *blockdevice.BlockDevice) {
	di := &blkid.DeviceIdentifier{DevPath: bd.DevPath}

	if len(bd.FSInfo.FileSystem) == 0 {
		bd.FSInfo.FileSystem = di.GetOnDiskFileSystem()
	}

	// if the host is CentOS 7, the `libblkid` version on host is `2.23`,
	// but the `PTUUID` tag was start to provide from `2.24`. This will cause
	// the udev cache fetched from host udevd will not contain env `ID_PART_TABLE_UUID`.
	// Fortunately, the `libblkid` version in our base container (ubuntu 16.04)
	// is `2.27.1`, will provide `PTUUID` tag, so we should fetch `PTUUID` from `blkid`.
	if len(bd.PartitionInfo.PartitionTableUUID) == 0 {
		bd.PartitionInfo.PartitionTableUUID = di.GetPartitionTableUUID()
	}

	// PARTUUID also is fetched using blkid, if udev is unable to get the data
	if len(bd.PartitionInfo.PartitionEntryUUID) == 0 {
		bd.PartitionInfo.PartitionEntryUUID = di.GetPartitionEntryUUID()
	}
}
