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
	"errors"
	"syscall"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

// UdevEventMessageChannel used to send event message
var UdevEventMessageChannel = make(chan controller.EventMessage)

// monitor contains udev and udevmonitor struct
type monitor struct {
	udev        *libudevwrapper.Udev
	udevMonitor *libudevwrapper.UdevMonitor
}

// newMonitor returns monitor struct in success
// we can get fd and monitor using this struct
func newMonitor() (*monitor, error) {
	udev, err := libudevwrapper.NewUdev()
	if err != nil {
		return nil, err
	}
	udevMonitor, err := udev.NewDeviceFromNetlink(libudevwrapper.UDEV_SOURCE)
	if err != nil {
		return nil, err
	}
	err = udevMonitor.AddSubsystemFilter(libudevwrapper.UDEV_SUBSYSTEM)
	if err != nil {
		return nil, err
	}
	err = udevMonitor.EnableReceiving()
	if err != nil {
		return nil, err
	}
	monitor := &monitor{
		udev:        udev,
		udevMonitor: udevMonitor,
	}
	return monitor, nil
}

// setup returns file descriptor value to monitor system
func (m *monitor) setup() (int, error) {
	return m.udevMonitor.GetFd()
}

// free frees udev and udevMonitor pointer
func (m *monitor) free() {
	if m.udev != nil {
		m.udev.UnrefUdev()
	}
	if m.udevMonitor != nil {
		m.udevMonitor.UdevMonitorUnref()
	}
}

// process get device info which is attached or detached
// generate one new event and sent it to udev probe
func (m *monitor) process(fd int) error {
	fds := &syscall.FdSet{}
	util.FD_ZERO(fds)
	util.FD_SET(fds, int(fd))
	ret, _ := syscall.Select(int(fd)+1, fds, nil, nil, nil)
	if ret <= 0 {
		return errors.New("unable to apply select call")
	}
	if !util.FD_ISSET(fds, int(fd)) {
		return errors.New("unable to set fd")
	}
	device, err := m.udevMonitor.ReceiveDevice()
	if err != nil {
		return err
	}
	// if device is not disk or partition, do not process it
	if !device.IsDisk() && !device.IsParitition() {
		device.UdevDeviceUnref()
		return nil
	}
	event := newEvent()
	event.process(device)
	event.send()
	return nil
}

//Monitor start monitoring on udev source
func Monitor() {
	monitor, err := newMonitor()
	if err != nil {
		klog.Error(err)
	}
	defer monitor.free()
	fd, err := monitor.setup()
	if err != nil {
		klog.Error(err)
	}
	for {
		err := monitor.process(fd)
		if err != nil {
			klog.Error(err)
		}
	}
}
