/*
Copyright 2018 The OpenEBS Author

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

package udevevent

import (
	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"k8s.io/klog"
)

// event contains EventMessage struct
type event struct {
	eventDetails controller.EventMessage
}

//newEvent returns a copy of event struct
func newEvent() *event {
	event := &event{
		eventDetails: controller.EventMessage{},
	}
	return event
}

// process takes udevdevice as input and generate event message
func (e *event) process(device *libudevwrapper.UdevDevice) {
	defer device.UdevDeviceUnref()
	diskInfo := make([]*blockdevice.BlockDevice, 0)
	uuid := device.GetUid()
	path := device.GetPath()
	action := device.GetAction()
	klog.Infof("processing new event for (%s)%s action type %s", path, uuid, action)
	deviceDetails := &blockdevice.BlockDevice{}
	deviceDetails.UUID = uuid
	deviceDetails.SysPath = device.GetSyspath()
	diskInfo = append(diskInfo, deviceDetails)
	e.eventDetails.Action = action
	e.eventDetails.Devices = diskInfo
}

// send sends event message to udev probe via channel
func (e *event) send() {
	UdevEventMessageChannel <- e.eventDetails
}
