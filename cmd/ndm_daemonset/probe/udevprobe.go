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

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/features"
	"github.com/openebs/node-disk-manager/pkg/sysfs"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/udevevent"
	"github.com/openebs/node-disk-manager/pkg/util"
	"golang.org/x/sync/semaphore"

	"k8s.io/klog"
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
		klog.Error("unable to configure", udevProbeName)
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
func newUdevProbeForFillDiskDetails(sysPath string) (*udevProbe, error) {
	udev, err := libudevwrapper.NewUdev()
	if err != nil {
		return nil, err
	}
	udevDevice, err := udev.NewDeviceFromSysPath(sysPath)
	if err != nil {
		return nil, err
	}
	udevProbe := &udevProbe{
		udev:       udev,
		udevDevice: udevDevice,
	}
	return udevProbe, nil
}

// Start setup udev probe listener and make a single scan of system
func (up *udevProbe) Start() {
	go up.listen()
	go udevevent.Monitor()
	probeEvent := newUdevProbe(up.controller)
	err := probeEvent.scan()
	if err != nil {
		klog.Errorf("error while scanning system for block devices, Error: %v", err)
	}
}

// Rescan syncs etcd and NDM
func Rescan(c *controller.Controller) error {
	udevProbe := newUdevProbe(c)
	defer udevProbe.free()
	err := udevProbe.scan()
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

var sem = semaphore.NewWeighted(1)

// scan scans system for block devices and send add event via channel
func (up *udevProbe) scan() error {

	// By using a semaphore, we ensure thread safety.
	if !sem.TryAcquire(1) {
		return errors.New("Scan is in progress")
	}
	defer sem.Release(1)

	if (up.udev == nil) || (up.udevEnumerate == nil) {
		return errors.New("unable to scan udev and udev enumerate is nil")
	}
	diskInfo := make([]*blockdevice.BlockDevice, 0)
	disksUid := make([]string, 0)
	err := up.udevEnumerate.AddSubsystemFilter(libudevwrapper.UDEV_SUBSYSTEM)
	if err != nil {
		return err
	}
	err = up.udevEnumerate.ScanDevices()
	if err != nil {
		return err
	}
	// everytime while performing the scan, we are re-initializing the
	// disk map of the system
	up.controller.BDHierarchy = make(blockdevice.Hierarchy)
	for l := up.udevEnumerate.ListEntry(); l != nil; l = l.GetNextEntry() {
		s := l.GetName()
		newUdevice, err := up.udev.NewDeviceFromSysPath(s)
		if err != nil {
			continue
		}
		if newUdevice.IsDisk() || newUdevice.IsParitition() {
			deviceDetails := &blockdevice.BlockDevice{}
			if features.FeatureGates.IsEnabled(features.GPTBasedUUID) {
				// WWN, Serial, PartitionTableUUID/GPTLabel, PartitionUUID, FileSystemUUID and DeviceType
				// are the fields we use to generate the UUID. These fields will be fetched
				// from the udev event itself. This is to guarantee that we do not need to rely
				// on any other probes to fill in those details which are critical for device identification.
				deviceDetails.DeviceAttributes.WWN = newUdevice.GetPropertyValue(libudevwrapper.UDEV_WWN)
				deviceDetails.DeviceAttributes.Serial = newUdevice.GetPropertyValue(libudevwrapper.UDEV_SERIAL)
				deviceDetails.PartitionInfo.PartitionTableUUID = newUdevice.GetPropertyValue(libudevwrapper.UDEV_PARTITION_TABLE_UUID)
				deviceDetails.PartitionInfo.PartitionEntryUUID = newUdevice.GetPropertyValue(libudevwrapper.UDEV_PARTITION_UUID)
				deviceDetails.FSInfo.FileSystemUUID = newUdevice.GetPropertyValue(libudevwrapper.UDEV_FS_UUID)
				deviceDetails.DMInfo.DMUUID = newUdevice.GetPropertyValue(libudevwrapper.UDEV_DM_UUID)
			} else {
				uuid := newUdevice.GetUid()
				disksUid = append(disksUid, uuid)
				deviceDetails.UUID = uuid
			}
			udevDeviceType := newUdevice.GetPropertyValue(libudevwrapper.UDEV_DEVTYPE)
			deviceDetails.SysPath = newUdevice.GetSyspath()
			deviceDetails.DevPath = newUdevice.GetPath()

			// log the details only if present, to avoid log flooding
			if deviceDetails.DeviceAttributes.WWN != "" {
				klog.V(4).Infof("device: %s, WWN: %s filled during udev scan",
					deviceDetails.DevPath, deviceDetails.DeviceAttributes.WWN)
			}
			if deviceDetails.DeviceAttributes.Serial != "" {
				klog.V(4).Infof("device: %s, Serial: %s filled during udev scan",
					deviceDetails.DevPath, deviceDetails.DeviceAttributes.Serial)
			}
			if deviceDetails.PartitionInfo.PartitionTableUUID != "" {
				klog.V(4).Infof("device: %s, PartitionTableUUID: %s filled during udev scan",
					deviceDetails.DevPath, deviceDetails.PartitionInfo.PartitionTableUUID)
			}
			if deviceDetails.PartitionInfo.PartitionEntryUUID != "" {
				klog.V(4).Infof("device: %s, PartitionEntryUUID: %s filled during udev scan",
					deviceDetails.DevPath, deviceDetails.PartitionInfo.PartitionEntryUUID)
			}
			if deviceDetails.FSInfo.FileSystemUUID != "" {
				klog.V(4).Infof("device: %s, FileSystemUUID: %s filled during udev scan",
					deviceDetails.DevPath, deviceDetails.FSInfo.FileSystemUUID)
			}

			sysfsDevice, err := sysfs.NewSysFsDeviceFromDevPath(deviceDetails.DevPath)
			// TODO if error occurs a rescan may be required
			if err != nil {
				klog.Errorf("could not get sysfs device for %s, err: %v", deviceDetails.DevPath, err)
			} else {
				// get the dependents of the block device
				// this is done by scanning sysfs
				dependents, err := sysfsDevice.GetDependents()
				// TODO if error occurs need to do a scan from the beginning
				if err != nil {
					klog.Errorf("error getting dependent devices for %s, err: %v", deviceDetails.DevPath, err)
				} else {
					deviceDetails.DependentDevices = dependents
					klog.Infof("Dependents of %s : %+v", deviceDetails.DevPath, dependents)
				}
				// the device type reported by udev will always be disk/partition. Using this info
				// and the entries from sysfs, the actual device type is found out.
				deviceType, err := sysfsDevice.GetDeviceType(udevDeviceType)
				if err != nil {
					klog.Errorf("could not get device type for %s, falling back to udev reported type: %s", deviceDetails.DevPath, udevDeviceType)
					deviceType = udevDeviceType
				}
				deviceDetails.DeviceAttributes.DeviceType = deviceType
				klog.Infof("Device: %s is of type: %s", deviceDetails.DevPath, deviceDetails.DeviceAttributes.DeviceType)
			}

			diskInfo = append(diskInfo, deviceDetails)
		}
		newUdevice.UdevDeviceUnref()
	}

	// when GPTBasedUUID is enabled, all the blockdevices will be made inactive initially.
	// after that each device that is detected by the probe will be marked as Active.
	up.controller.DeactivateStaleBlockDeviceResource(disksUid)
	eventDetails := controller.EventMessage{
		Action:  libudevwrapper.UDEV_ACTION_ADD,
		Devices: diskInfo,
	}
	udevevent.UdevEventMessageChannel <- eventDetails
	return nil
}

// fillDiskDetails fills details in diskInfo struct using probe information
func (up *udevProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	udevDevice, err := newUdevProbeForFillDiskDetails(blockDevice.SysPath)
	if err != nil {
		klog.Errorf("%s : %s", blockDevice.SysPath, err)
		return
	}
	udevDiskDetails := udevDevice.udevDevice.DiskInfoFromLibudev()
	defer udevDevice.free()
	blockDevice.DevPath = udevDiskDetails.Path
	blockDevice.DeviceAttributes.Model = udevDiskDetails.Model
	blockDevice.DeviceAttributes.WWN = udevDiskDetails.WWN
	blockDevice.DeviceAttributes.Serial = udevDiskDetails.Serial
	blockDevice.DeviceAttributes.Vendor = udevDiskDetails.Vendor
	blockDevice.DeviceAttributes.IDType = udevDiskDetails.IDType

	blockDevice.DMInfo.DevMapperPath = udevDiskDetails.DMPath

	// log only if details are present to prevent log flooding
	if blockDevice.DeviceAttributes.Model != "" {
		klog.V(4).Infof("device: %s, Model: %s filled by udev probe",
			blockDevice.DevPath, blockDevice.DeviceAttributes.Model)
	}
	if blockDevice.DeviceAttributes.WWN != "" {
		klog.V(4).Infof("device: %s, WWN: %s filled by udev probe",
			blockDevice.DevPath, blockDevice.DeviceAttributes.WWN)
	}
	if blockDevice.DeviceAttributes.Serial != "" {
		klog.V(4).Infof("device: %s, Serial: %s filled by udev probe",
			blockDevice.DevPath, blockDevice.DeviceAttributes.Serial)
	}
	if blockDevice.DeviceAttributes.Vendor != "" {
		klog.V(4).Infof("device: %s, Vendor: %s filled by udev probe",
			blockDevice.DevPath, blockDevice.DeviceAttributes.Vendor)
	}
	if blockDevice.DeviceAttributes.IDType != "" {
		klog.V(4).Infof("device: %s, IDType: %s filled by udev probe",
			blockDevice.DevPath, blockDevice.DeviceAttributes.IDType)
	}

	if len(udevDiskDetails.ByIdDevLinks) != 0 {
		blockDevice.DevLinks = append(blockDevice.DevLinks, blockdevice.DevLink{
			Kind:  libudevwrapper.BY_ID_LINK,
			Links: udevDiskDetails.ByIdDevLinks,
		})
	}

	if len(udevDiskDetails.ByPathDevLinks) != 0 {
		blockDevice.DevLinks = append(blockDevice.DevLinks, blockdevice.DevLink{
			Kind:  libudevwrapper.BY_PATH_LINK,
			Links: udevDiskDetails.ByPathDevLinks,
		})
	}

	// filesystem info of the attached device. Only filesystem data will be filled in the struct,
	// as the mountpoint related information will be filled in by the mount probe
	blockDevice.FSInfo.FileSystem = udevDiskDetails.FileSystem

	blockDevice.PartitionInfo.PartitionTableType = udevDiskDetails.PartitionTableType

	// if this is a partition, partition number and partition UUID need to be filled
	if udevDiskDetails.DiskType == blockdevice.BlockDeviceTypePartition {
		blockDevice.PartitionInfo.PartitionNumber = udevDiskDetails.PartitionNumber
	}
}

// listen listens for event message over UdevEventMessages channel
// when it gets event via channel it transfer to event handler
// this function is blocking function better to use it in a routine.
func (up *udevProbe) listen() {
	if up.controller == nil {
		klog.Error("unable to setup udev probe listener controller object is nil")
		return
	}
	probeEvent := ProbeEvent{
		Controller: up.controller,
	}
	klog.Info("starting udev probe listener")
	for {
		msg := <-udevevent.UdevEventMessageChannel
		switch msg.Action {
		case string(AttachEA):
			probeEvent.addBlockDeviceEvent(msg)
		case string(DetachEA):
			probeEvent.deleteBlockDeviceEvent(msg)
		case string(MountEA):
			probeEvent.changeBlockDeviceEvent(msg)
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
