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
	. "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/mount"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
	"strings"
)

// mountProbe contains required variables for populating diskInfo
type mountProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller      *controller.Controller
	MountIdentifier *mount.Identifier
}

const (
	mountProbePriority = 5
	mountConfigKey     = "mount-probe"
	k8sLocalVolumePath = "kubernetes.io/local-volume"
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
		pi:         &mountProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the probe (smart probe in this case)
	newRegisterProbe.register()
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

// Start is mainly used for one time activities such as monitoring.
// It is a part of probe interface but here we does not require to perform
// such activities, hence empty implementation
func (mp *mountProbe) Start() {}

// FillBlockDeviceDetails fills details in diskInfo struct using information it gets from probe
func (mp *mountProbe) FillBlockDeviceDetails(blockDevice *BlockDevice) {
	if blockDevice.DevPath == "" {
		klog.Error("mountIdentifier is found empty, mount probe will not fetch mount information.")
		return
	}
	mountProbe := newMountProbe(blockDevice.DevPath)
	basicMountInfo, err := mountProbe.MountIdentifier.DeviceBasicMountInfo()
	if err != nil {
		klog.Error(err)
		return
	}
	// if mount point contains kubernetes.io/local-volume it means that the device is mounted
	// by kubelet. In that case we ignore the mountpoint.
	if strings.Contains(basicMountInfo.MountPoint, k8sLocalVolumePath) {
		return
	}
	blockDevice.FSInfo.MountPoint = append(blockDevice.FSInfo.MountPoint, basicMountInfo.MountPoint)
	if blockDevice.FSInfo.FileSystem == "" {
		blockDevice.FSInfo.FileSystem = basicMountInfo.FileSystem
	}
}
