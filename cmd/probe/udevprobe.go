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
	"errors"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/cmd/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/udevevent"
	"github.com/openebs/node-disk-manager/pkg/util"
)

const (
	udevProbePriority = 1
	udevConfigKey     = "udev-probe"
)

var (
	udevProbeName  = "udev probe"
	udevProbeState = defaultEnabled
)

// udevProbeRegister contains registration process of udev probe
var udevProbeRegister = func() {
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		glog.Error("unable to configure", udevProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == udevConfigKey {
				udevProbeName = probeConfig.Name
				udevProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   udevProbePriority,
		name:       udevProbeName,
		state:      udevProbeState,
		pi:         newUdevProbe(ctrl),
		controller: ctrl,
	}
	newRegisterProbe.register()
}

// udevProbe contains require variables for scan , populate diskInfo and push
// resource in etcd
type udevProbe struct {
	controller    *controller.Controller
	udev          *libudevwrapper.Udev
	udevDevice    *libudevwrapper.UdevDevice
	udevEnumerate *libudevwrapper.UdevEnumerate
}

// newUdevProbe returns udevProbe struct which helps to setup probe listen and scan
// system it contains copy of udev, udevEnumerate struct use defer free() in caller
// function to free c pointer memory.
func newUdevProbe(c *controller.Controller) *udevProbe {
	udev, err := libudevwrapper.NewUdev()
	if err != nil {
		return nil
	}
	udevEnumerate, err := udev.NewUdevEnumerate()
	if err != nil {
		return nil
	}
	udevProbe := &udevProbe{
		controller:    c,
		udev:          udev,
		udevEnumerate: udevEnumerate,
	}
	return udevProbe
}

// newUdevProbeForFillDiskDetails returns udevProbe struct which helps populate diskInfo struct.
// it contains copy of udevDevice struct to populate diskInfo use defer free in caller function
// to free c pointer memory
func newUdevProbeForFillDiskDetails(sysPath string) *udevProbe {
	udev, err := libudevwrapper.NewUdev()
	if err != nil {
		return nil
	}
	udevDevice, err := udev.NewDeviceFromSysPath(sysPath)
	if err != nil {
		return nil
	}
	udevProbe := &udevProbe{
		udev:       udev,
		udevDevice: udevDevice,
	}
	return udevProbe
}

// Start setup udev probe listener and make a single scan of system
func (up *udevProbe) Start() {
	go up.listen()
	go udevevent.Monitor()
	probeEvent := newUdevProbe(up.controller)
	probeEvent.scan()
}

// scan scans system for disks and send add event via channel
func (up *udevProbe) scan() error {
	if (up.udev == nil) || (up.udevEnumerate == nil) {
		return errors.New("unable to scan udev and udev enumerate is nil")
	}
	diskInfo := make([]*controller.DiskInfo, 0)
	disksUid := make([]string, 0)
	err := up.udevEnumerate.AddSubsystemFilter(libudevwrapper.UDEV_SUBSYSTEM)
	if err != nil {
		return err
	}
	err = up.udevEnumerate.ScanDevices()
	if err != nil {
		return err
	}
	for l := up.udevEnumerate.ListEntry(); l != nil; l = l.GetNextEntry() {
		s := l.GetName()
		newUdevice, err := up.udev.NewDeviceFromSysPath(s)
		if err != nil {
			continue
		}
		if newUdevice.IsDisk() {
			uuid := up.controller.GenerateUuid(newUdevice)
			disksUid = append(disksUid, uuid)
			deviceDetails := &controller.DiskInfo{}
			deviceDetails.ProbeIdentifiers.Uuid = uuid
			deviceDetails.ProbeIdentifiers.UdevIdentifier = newUdevice.GetSyspath()
			diskInfo = append(diskInfo, deviceDetails)
		}
		newUdevice.UdevDeviceUnref()
	}
	up.controller.DeactivateStaleDiskResource(disksUid)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: diskInfo,
	}
	udevevent.UdevEventMessageChannel <- eventDetails
	return nil
}

// fillDiskDetails filles details in diskInfo struct using probe information
func (up *udevProbe) FillDiskDetails(d *controller.DiskInfo) {
	udevDevice := newUdevProbeForFillDiskDetails(d.ProbeIdentifiers.UdevIdentifier)
	udevDiskDetails := udevDevice.udevDevice.DiskInfoFromLibudev()
	defer udevDevice.free()
	d.ProbeIdentifiers.SmartIdentifier = udevDiskDetails.Path
	d.ProbeIdentifiers.SeachestIdentifier = udevDiskDetails.Path
	d.Model = udevDiskDetails.Model
	d.Path = udevDiskDetails.Path
	d.Serial = udevDiskDetails.Serial
	d.Vendor = udevDiskDetails.Vendor
	d.ByIdDevLinks = udevDiskDetails.ByIdDevLinks
	d.ByPathDevLinks = udevDiskDetails.ByPathDevLinks
}

// listen listens for event message over UdevEventMessages channel
// when it gets event via channel it transfer to event handler
// this function is blocking function better to use it in a routine.
func (up *udevProbe) listen() {
	if up.controller == nil {
		glog.Error("unable to setup udev probe listener controller object is nil")
		return
	}
	probeEvent := ProbeEvent{
		Controller: up.controller,
	}
	glog.Info("starting udev probe listener")
	for {
		msg := <-udevevent.UdevEventMessageChannel
		switch msg.Action {
		case string(AttachEA):
			probeEvent.addDiskEvent(msg)
		case string(DetachEA):
			probeEvent.deleteDiskEvent(msg)
		}
	}
}

// free frees c pointers if it is not null
func (up *udevProbe) free() {
	if up.udev != nil {
		up.udev.UnrefUdev()
	}
	if up.udevDevice != nil {
		up.udevDevice.UdevDeviceUnref()
	}
	if up.udevEnumerate != nil {
		up.udevEnumerate.UnrefUdevEnumerate()
	}
}
