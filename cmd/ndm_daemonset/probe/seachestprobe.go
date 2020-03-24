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
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/seachest"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

// seachest contains required variables for populating diskInfo
type seachestProbe struct {
	// Every new probe needs a controller object to register itself.
	// Here Controller consists of Clientset, kubeClientset, probes, etc which is used to
	// create, update, delete, deactivate the disk resources or list the probes already registered.
	Controller         *controller.Controller
	SeachestIdentifier *seachest.Identifier
}

const (
	seachestConfigKey     = "seachest-probe"
	seachestProbePriority = 4
)

var (
	seachestProbeName  = "seachest probe"
	seachestProbeState = defaultEnabled
)

// init is used to get a controller object and then register itself
var seachestProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", seachestProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == seachestConfigKey {
				seachestProbeName = probeConfig.Name
				seachestProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   seachestProbePriority,
		name:       seachestProbeName,
		state:      seachestProbeState,
		pi:         &seachestProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the probe (seachest probe in this case)
	newRegisterProbe.register()
}

// newSeachestProbe returns seachestProbe struct which helps populate diskInfo struct
// with the basic disk details such as logical size, firmware revision, etc
func newSeachestProbe(devPath string) *seachestProbe {
	seachestIdentifier := &seachest.Identifier{
		DevPath: devPath,
	}
	seachestProbe := &seachestProbe{
		SeachestIdentifier: seachestIdentifier,
	}
	return seachestProbe
}

// Start is mainly used for one time activities such as monitoring.
// It is a part of probe interface but here we does not require to perform
// such activities, hence empty implementation
func (scp *seachestProbe) Start() {}

// fillDiskDetails fills details in diskInfo struct using information it gets from probe
func (scp *seachestProbe) FillDiskDetails(d *controller.DiskInfo) {
	if d.ProbeIdentifiers.SeachestIdentifier == "" {
		klog.Error("seachestIdentifier is found empty, seachest probe will not fill disk details.")
		return
	}

	seachestProbe := newSeachestProbe(d.ProbeIdentifiers.SeachestIdentifier)
	driveInfo, err := seachestProbe.SeachestIdentifier.SeachestBasicDiskInfo()
	if err != 0 {
		klog.Error(err)
		return
	}

	if d.Path == "" {
		d.Path = seachestProbe.SeachestIdentifier.GetPath(driveInfo)
		klog.V(4).Infof("Path:%s filled by seachest.", d.Path)
	}

	if d.NodeAttributes[controller.HostNameKey] == "" {
		d.NodeAttributes[controller.HostNameKey] = seachestProbe.SeachestIdentifier.GetHostName(driveInfo)
		klog.V(4).Infof("Disk: %s NodeAttribute:%s filled by seachest.", d.Path, d.NodeAttributes[controller.HostNameKey])
	}

	if d.Model == "" {
		d.Model = seachestProbe.SeachestIdentifier.GetModelNumber(driveInfo)
		klog.V(4).Infof("Disk: %s Model:%s filled by seachest.", d.Path, d.Model)
	}

	if d.Uuid == "" {
		d.Uuid = seachestProbe.SeachestIdentifier.GetUuid(driveInfo)
		klog.V(4).Infof("Disk: %s Uuid:%s filled by seachest.", d.Path, d.Uuid)
	}

	if d.Capacity == 0 {
		d.Capacity = seachestProbe.SeachestIdentifier.GetCapacity(driveInfo)
		klog.V(4).Infof("Disk: %s Capacity:%d filled by seachest.", d.Path, d.Capacity)
	}

	if d.Serial == "" {
		d.Serial = seachestProbe.SeachestIdentifier.GetSerialNumber(driveInfo)
		klog.V(4).Infof("Disk: %s Serial:%s filled by seachest.", d.Path, d.Serial)
	}

	if d.Vendor == "" {
		d.Vendor = seachestProbe.SeachestIdentifier.GetVendorID(driveInfo)
		klog.V(4).Infof("Disk: %s Vendor:%s filled by seachest.", d.Path, d.Vendor)
	}

	if d.FirmwareRevision == "" {
		d.FirmwareRevision = seachestProbe.SeachestIdentifier.GetFirmwareRevision(driveInfo)
		klog.V(4).Infof("Disk: %s FirmwareRevision:%s filled by seachest.", d.Path, d.FirmwareRevision)
	}

	if d.LogicalSectorSize == 0 {
		d.LogicalSectorSize = seachestProbe.SeachestIdentifier.GetLogicalSectorSize(driveInfo)
		klog.V(4).Infof("Disk: %s LogicalBlockSize:%d filled by seachest.", d.Path, d.LogicalSectorSize)
	}

	if d.PhysicalSectorSize == 0 {
		d.PhysicalSectorSize = seachestProbe.SeachestIdentifier.GetPhysicalSectorSize(driveInfo)
		klog.V(4).Infof("Disk: %s PhysicalBlockSize:%d filled by seachest.", d.Path, d.PhysicalSectorSize)
	}

	if d.RotationRate == 0 {
		d.RotationRate = seachestProbe.SeachestIdentifier.GetRotationRate(driveInfo)
		klog.V(4).Infof("Disk: %s RotationRate:%d filled by seachest.", d.Path, d.RotationRate)
	}

	if d.DriveType == "" {
		d.DriveType = seachestProbe.SeachestIdentifier.DriveType(driveInfo)
		klog.V(4).Infof("Disk: %s DriveType:%s filled by seachest.", d.Path, d.DriveType)
	}

	if d.TotalBytesRead == 0 {
		d.TotalBytesRead = seachestProbe.SeachestIdentifier.GetTotalBytesRead(driveInfo)
		klog.V(4).Infof("Disk: %s TotalBytesRead:%d filled by seachest.", d.Path, d.TotalBytesRead)
	}

	if d.TotalBytesWritten == 0 {
		d.TotalBytesWritten = seachestProbe.SeachestIdentifier.GetTotalBytesWritten(driveInfo)
		klog.V(4).Infof("Disk: %s TotalBytesWritten:%d filled by seachest.", d.Path, d.TotalBytesWritten)
	}

	if d.DeviceUtilizationRate == 0 {
		d.DeviceUtilizationRate = seachestProbe.SeachestIdentifier.GetDeviceUtilizationRate(driveInfo)
		klog.V(4).Infof("Disk: %s DeviceUtilizationRate:%f filled by seachest.", d.Path, d.DeviceUtilizationRate)
	}

	if d.PercentEnduranceUsed == 0 {
		d.PercentEnduranceUsed = seachestProbe.SeachestIdentifier.GetPercentEnduranceUsed(driveInfo)
		klog.V(4).Infof("Disk: %s PercentEnduranceUsed:%f filled by seachest.", d.Path, d.PercentEnduranceUsed)
	}

	d.TemperatureInfo.TemperatureDataValid = seachestProbe.
		SeachestIdentifier.GetTemperatureDataValidStatus(driveInfo)
	klog.V(4).Infof("Disk: %s TemperatureDataValid:%t filled by seachest.",
		d.Path, d.TemperatureInfo.TemperatureDataValid)

	if d.TemperatureInfo.TemperatureDataValid == true {
		d.TemperatureInfo.CurrentTemperature = seachestProbe.
			SeachestIdentifier.GetCurrentTemperature(driveInfo)

		klog.V(4).Infof("Disk: %s CurrentTemperature:%d filled by seachest.",
			d.Path, d.TemperatureInfo.CurrentTemperature)

		d.TemperatureInfo.HighestValid = seachestProbe.
			SeachestIdentifier.GetHighestValid(driveInfo)

		klog.V(4).Infof("Disk: %s HighestValid:%t filled by seachest.",
			d.Path, d.TemperatureInfo.HighestValid)

		d.TemperatureInfo.HighestTemperature = seachestProbe.
			SeachestIdentifier.GetHighestTemperature(driveInfo)

		klog.V(4).Infof("Disk: %s HighestTemperature:%d filled by seachest.",
			d.Path, d.TemperatureInfo.HighestTemperature)

		d.TemperatureInfo.LowestValid = seachestProbe.
			SeachestIdentifier.GetLowestValid(driveInfo)

		klog.V(4).Infof("Disk: %s LowestValid:%t filled by seachest.",
			d.Path, d.TemperatureInfo.LowestValid)

		d.TemperatureInfo.LowestTemperature = seachestProbe.
			SeachestIdentifier.GetLowestTemperature(driveInfo)

		klog.V(4).Infof("Disk: %s LowestTemperature:%d filled by seachest.",
			d.Path, d.TemperatureInfo.LowestTemperature)
	}
}
