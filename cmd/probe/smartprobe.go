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
	"github.com/openebs/node-disk-manager/pkg/smart"
)

// smartProbe contains required variables for populating diskInfo
type smartProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller      *controller.Controller
	SmartIdentifier *smart.Identifier
}

const (
	smartProbeName     = "smart probe"
	smartProbeState    = defaultEnabled
	smartProbePriority = 2
)

// init is used to get a controller object and then register itself
func init() {
	go func() {
		// Get a controller object
		ctrl := <-controller.ControllerBroadcastChannel
		if ctrl == nil {
			glog.Error("unable to configure", smartProbeName)
			return
		}
		var pi controller.ProbeInterface = &smartProbe{Controller: ctrl}
		newPrgisterProbe := &registerProbe{
			priority:       smartProbePriority,
			probeName:      smartProbeName,
			probeState:     smartProbeState,
			probeInterface: pi,
			controller:     ctrl,
		}
		// Here we register the probe (smart probe in this case)
		newPrgisterProbe.register()
	}()
}

// newSmartProbe returns smartProbe struct which helps populate diskInfo struct
// with the basic disk details such as logical size, firmware revision, etc
func newSmartProbe(devPath string) *smartProbe {
	smartIdentifier := &smart.Identifier{
		DevPath: devPath,
	}
	smartProbe := &smartProbe{
		SmartIdentifier: smartIdentifier,
	}
	return smartProbe
}

// Start is mainly used for one time activities such as monitoring.
// It is a part of probe interface but here we does not require to perform
// such activities, hence empty implementation
func (sp *smartProbe) Start() {}

// fillDiskDetails fills details in diskInfo struct using information it gets from probe
func (sp *smartProbe) FillDiskDetails(d *controller.DiskInfo) {
	if d.ProbeIdentifiers.SmartIdentifier == "" {
		glog.Error("smartIdentifier is found empty, smart probe will not fill disk details.")

		return
	}
	smartProbe := newSmartProbe(d.ProbeIdentifiers.SmartIdentifier)
	deviceBasicSCSIInfo, err := smartProbe.SmartIdentifier.SCSIBasicDiskInfo()
	if len(err) != 0 {
		glog.Error(err)
	}

	d.SPCVersion = deviceBasicSCSIInfo.SPCVersion
	d.FirmwareRevision = deviceBasicSCSIInfo.FirmwareRevision
	d.Capacity = deviceBasicSCSIInfo.Capacity
	d.LogicalSectorSize = deviceBasicSCSIInfo.LBSize

}
