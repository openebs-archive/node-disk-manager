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
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
)

type EventAction string

const (
	AttachEA EventAction = libudevwrapper.UDEV_ACTION_ADD    // AttachEA is atach disk event name
	DetachEA EventAction = libudevwrapper.UDEV_ACTION_REMOVE // DetachEA is detach disk event name
)

// ProbeEvent staruct contain a copy of controller it will update disk resources
type ProbeEvent struct {
	Controller *controller.Controller
}

// addDiskEvent fill disk details from different probes and push it to etcd
func (pe *ProbeEvent) addDiskEvent(msg controller.EventMessage) {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		glog.Error(err)
		pe.initOrErrorEvent()
		return
	}
	for _, diskDetails := range msg.Devices {
		glog.Info("processing data for ", diskDetails.ProbeIdentifiers.Uuid)
		diskDetails.HostName = pe.Controller.HostName
		diskDetails.Uuid = diskDetails.ProbeIdentifiers.Uuid
		probes := pe.Controller.ListProbe()
		for _, probe := range probes {
			glog.Info("disk details filled by ", probe.ProbeName)
			probe.FillDiskDetails(diskDetails)
		}
		diskApi := diskDetails.ToDisk()
		oldDr := pe.Controller.GetExistingResource(diskList, diskDetails.ProbeIdentifiers.Uuid)
		if oldDr != nil {
			pe.Controller.UpdateDisk(diskApi, oldDr)
			continue
		}
		pe.Controller.CreateDisk(diskApi)
	}
}

// deleteDiskEvent deactivate disk resource using uuid from etcd
func (pe *ProbeEvent) deleteDiskEvent(msg controller.EventMessage) {
	diskList, err := pe.Controller.ListDiskResource()
	if err != nil {
		glog.Error(err)
		pe.initOrErrorEvent()
		return
	}
	mismatch := false
	// set mismatch = true when one disk is removed from node and
	// entry related that disk not present in etcd in that case it
	// again rescan full system and update etcd accordingly.
	for _, diskDetails := range msg.Devices {
		oldDr := pe.Controller.GetExistingResource(diskList, diskDetails.ProbeIdentifiers.Uuid)
		if oldDr == nil {
			mismatch = true
			continue
		}
		pe.Controller.DeactivateDisk(*oldDr)
	}
	if mismatch {
		pe.initOrErrorEvent()
	}
}

// initOrErrorEvent rescan system and update disk resource this is
// used for initial setup and when any uid mismatch or error occurred.
func (pe *ProbeEvent) initOrErrorEvent() {
	udevProbe := newUdevProbe(pe.Controller)
	err := udevProbe.scan()
	if err != nil {
		glog.Error(err)
	}
}
