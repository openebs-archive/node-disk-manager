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

	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
)

type UdevEventType uint

// monitor contains udev and udevmonitor struct
type monitor struct {
	udev        *libudevwrapper.Udev
	udevMonitor *libudevwrapper.UdevMonitor
}

type UdevEvent struct {
	*libudevwrapper.UdevDevice
	eventType UdevEventType
}

type Subscription struct {
	targetChannel   chan UdevEvent
	subscribedTypes []UdevEventType
}

const (
	EventTypeAdd UdevEventType = iota
	EventTypeRemove
	EventTypeChange
)

var subscriptions []*Subscription

var ErrInvalidSubscription = errors.New("invailid subscription")

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
	var eventType UdevEventType
	switch device.GetAction() {
	case libudevwrapper.UDEV_ACTION_ADD:
		eventType = EventTypeAdd
	case libudevwrapper.UDEV_ACTION_REMOVE:
		eventType = EventTypeRemove
	default:
		eventType = EventTypeChange
	}
	event := UdevEvent{device, eventType}
	dispatchEvent(event)
	return nil
}

func dispatchEvent(event UdevEvent) {
	for _, sub := range subscriptions {
		hasType := false
		for _, eventType := range sub.subscribedTypes {
			if eventType == event.eventType {
				hasType = true
				break
			}
		}
		if hasType {
			sub.targetChannel <- event
		}
	}
}

//Monitor start monitoring on udev source
func Monitor() <-chan error {
	errChan := make(chan error)
	go func() {
		monitor, err := newMonitor()
		if err != nil {
			errChan <- err
		}
		defer monitor.free()
		fd, err := monitor.setup()
		if err != nil {
			errChan <- err
		}
		for {
			err := monitor.process(fd)
			if err != nil {
				errChan <- err
			}
		}
	}()
	return errChan
}

func Subscribe(eventTypes ...UdevEventType) *Subscription {
	if len(eventTypes) == 0 {
		eventTypes = []UdevEventType{EventTypeAdd, EventTypeRemove, EventTypeChange}
	}
	subscription := Subscription{
		targetChannel:   make(chan UdevEvent),
		subscribedTypes: eventTypes,
	}
	subscriptions = append(subscriptions, &subscription)
	return &subscription
}

func Unsubscribe(sub *Subscription) error {
	if sub == nil || sub.targetChannel == nil || sub.subscribedTypes == nil {
		return ErrInvalidSubscription
	}
	var deleteIndex = -1
	for idx, subscription := range subscriptions {
		if subscription == sub {
			deleteIndex = idx
		}
	}
	close(sub.targetChannel)
	if deleteIndex == len(subscriptions)-1 {
		subscriptions = subscriptions[:deleteIndex]
	} else if deleteIndex == 0 {
		subscriptions = subscriptions[1:]
	} else {
		subscriptions = append(subscriptions[:deleteIndex], subscriptions[deleteIndex+1:]...)
	}
	return nil
}

func (s *Subscription) Events() <-chan UdevEvent {
	return s.targetChannel
}

func (u UdevEvent) GetType() UdevEventType {
	return u.eventType
}

func (u UdevEvent) GetAction() UdevEventType {
	return u.eventType
}

func (uevt UdevEventType) String() string {
	switch uevt {
	case EventTypeAdd:
		return "add"
	case EventTypeRemove:
		return "remove"
	case EventTypeChange:
		return "change"
	default:
		return "unknown"
	}
}
