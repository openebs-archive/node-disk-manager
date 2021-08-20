/*
Copyright 2019 The OpenEBS Authors

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
	"k8s.io/klog"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/epoll"
	"github.com/openebs/node-disk-manager/pkg/features"
	"github.com/openebs/node-disk-manager/pkg/mount"
	"github.com/openebs/node-disk-manager/pkg/mount/libmount"
	"github.com/openebs/node-disk-manager/pkg/util"
)

// mountProbe contains required variables for populating diskInfo
type mountProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller      *controller.Controller
	MountIdentifier *mount.Identifier
	epoll           *epoll.Epoll
	destination     chan controller.EventMessage
	mountsFileName  string
	mountTable      *libmount.MountTab
}

const (
	mountProbePriority = 4
	mountConfigKey     = "mount-probe"
)

var (
	mountProbeName  = "mount probe"
	mountProbeState = defaultEnabled
)

// init is used to get a controller object and then register itself
var mountProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", mountProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == mountConfigKey {
				mountProbeName = probeConfig.Name
				mountProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   mountProbePriority,
		name:       mountProbeName,
		state:      mountProbeState,
		pi:         newMountProbeForRegistration(ctrl),
		controller: ctrl,
	}
	// Here we register the probe (mount probe in this case)
	newRegisterProbe.register()
}

// newMountProbeForRegistration returns mountprobe struct which helps
// register the probe and start mount-point and fs change detection loop
func newMountProbeForRegistration(c *controller.Controller) *mountProbe {
	return &mountProbe{
		Controller:     c,
		mountsFileName: mount.HostMountsFilePath,
		destination:    controller.EventMessageChannel,
	}
}

// newMountProbe returns mountProbe struct which helps populate diskInfo struct
// with the mount related details like mountpoint
func newMountProbe(devPath string) *mountProbe {
	mountIdentifier := &mount.Identifier{
		DevPath: devPath,
	}
	mountProbe := &mountProbe{
		MountIdentifier: mountIdentifier,
	}
	return mountProbe
}

// Start initializes moutprobe and sets up necessary watchers
// for mount change detection
func (mp *mountProbe) Start() {
	if !features.FeatureGates.
		IsEnabled(features.ChangeDetection) {
		return
	}
	if err := mp.setupEpoll(); err != nil {
		klog.Errorf("failed to setup epoll: %v", err)
		return
	}
	mt, err := mp.newMountTable()
	if err != nil {
		klog.Errorf("failed to generate mount table")
		return
	}
	mp.mountTable = mt
	go mp.listen()
}

// FillBlockDeviceDetails fills details in diskInfo struct using information it gets from probe
func (mp *mountProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	if blockDevice.DevPath == "" {
		klog.Error("mountIdentifier is found empty, mount probe will not fetch mount information.")
		return
	}
	mountProbe := newMountProbe(blockDevice.DevPath)
	basicMountInfo, err := mountProbe.MountIdentifier.DeviceBasicMountInfo(mount.HostMountsFilePath)
	if err != nil {
		if err == mount.ErrAttributesNotFound {
			klog.Infof("no mount point found for %s. clearing mount points if any",
				blockDevice.DevPath)
			blockDevice.FSInfo.MountPoint = nil
			return
		}
		klog.Error(err)
		return
	}

	blockDevice.FSInfo.MountPoint = basicMountInfo.MountPoint
	if blockDevice.FSInfo.FileSystem == "" {
		blockDevice.FSInfo.FileSystem = basicMountInfo.FileSystem
	}
}

func (mp *mountProbe) setupEpoll() error {
	ep, err := epoll.New()
	if err != nil {
		return err
	}
	mp.epoll = &ep
	return ep.AddWatcher(epoll.Watcher{
		FileName:   mp.mountsFileName,
		EventTypes: []epoll.EventType{epoll.EPOLLERR, epoll.EPOLLPRI},
	})
}

func (mp *mountProbe) listen() {
	eventCh, err := mp.epoll.Start()
	if err != nil {
		klog.Errorf("error while starting epoll: %v", err)
		return
	}
	defer mp.epoll.Stop()
	klog.Info("started mount change detection loop")
	defaultMsg := controller.EventMessage{
		Action:          string(ChangeEA),
		Devices:         nil,
		AllBlockDevices: true,
	}

	for range eventCh {
		// regenerate mounts table and get the changes
		newMountTable, err := mp.newMountTable()
		if err != nil {
			klog.Error("failed to generate mounts table.")
			mp.destination <- defaultMsg
		}
		mtDiff := libmount.GenerateDiff(mp.mountTable, newMountTable)
		mp.mountTable = newMountTable
		mp.processDiff(mtDiff)
	}
}

func (mp *mountProbe) newMountTable() (*libmount.MountTab, error) {
	return libmount.NewMountTab(libmount.FromFile(mp.mountsFileName,
		libmount.MntFmtFstab),
		libmount.WithAllowFilter(libmount.SourceContainsFilter("/dev/")),
		libmount.WithDenyFilter(libmount.SourceFilter("overlay")),
		libmount.WithDenyFilter(libmount.TargetContainsFilter("/var/lib/kubelet/pod")),
		libmount.WithDenyFilter(libmount.TargetContainsFilter("/var/lib/docker")),
		libmount.WithDenyFilter(libmount.TargetContainsFilter("/run/docker")))
}

func (mp *mountProbe) processDiff(diff libmount.MountTabDiff) {
	devices := make([]*blockdevice.BlockDevice, 0)
	changedDevices := diff.ListSources()
	for _, dev := range changedDevices {
		bd := new(blockdevice.BlockDevice)
		bd.DevPath = dev
		devices = append(devices, bd)
	}

	mp.destination <- controller.EventMessage{
		Action:          string(ChangeEA),
		Devices:         devices,
		RequestedProbes: []string{mountProbeName},
	}

}
