/*
Copyright 2020 The OpenEBS Authors

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
	"os"
	"strings"
	"syscall"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller"
	"github.com/openebs/node-disk-manager/pkg/blkid"
	"github.com/openebs/node-disk-manager/pkg/spdk"
	libudevwrapper "github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"

	"k8s.io/klog"
)

// TODO give some good name to this probe
type samplingProbe struct {
	Controller      *controller.Controller
	BlkidIdentifier *blkid.DeviceIdentifier
}

const (
	samplingConfigKey     = "sampling-probe"
	samplingProbePriority = 4

	k8sLocalVolumePath1 = "kubernetes.io/local-volume"
	k8sLocalVolumePath2 = "kubernetes.io~local-volume"
	zfsFileSystemLabel  = "zfs_member"
)

var (
	samplingProbeName  = "sampling probe"
	samplingProbeState = defaultEnabled
)

var samplingProbeRegister = func() {
	// Get a controller object
	ctrl := <-controller.ControllerBroadcastChannel
	if ctrl == nil {
		klog.Error("unable to configure", samplingProbeName)
		return
	}
	if ctrl.NDMConfig != nil {
		for _, probeConfig := range ctrl.NDMConfig.ProbeConfigs {
			if probeConfig.Key == samplingConfigKey {
				samplingProbeName = probeConfig.Name
				samplingProbeState = util.CheckTruthy(probeConfig.State)
				break
			}
		}
	}
	newRegisterProbe := &registerProbe{
		priority:   samplingProbePriority,
		name:       samplingProbeName,
		state:      samplingProbeState,
		pi:         &samplingProbe{Controller: ctrl},
		controller: ctrl,
	}
	// Here we register the sampling probe
	newRegisterProbe.register()
}

func newSamplingProbe(devPath string) *samplingProbe {
	samplingProbe := &samplingProbe{
		BlkidIdentifier: &blkid.DeviceIdentifier{
			DevPath: devPath,
		},
	}
	return samplingProbe
}

func (sp *samplingProbe) Start() {}

func (sp *samplingProbe) FillBlockDeviceDetails(blockDevice *blockdevice.BlockDevice) {
	if blockDevice.DevPath == "" {
		klog.Errorf("sampling identifier found empty, sampling probe will not fetch information")
		return
	}

	// checking for local PV on the device
	for _, mountPoint := range blockDevice.FSInfo.MountPoint {
		if strings.Contains(mountPoint, k8sLocalVolumePath1) ||
			strings.Contains(mountPoint, k8sLocalVolumePath2) {
			blockDevice.DevUse.InUse = true
			blockDevice.DevUse.UsedBy = blockdevice.LocalPV
			klog.V(4).Infof("device: %s Used by: %s filled by sampling probe", blockDevice.DevPath, blockDevice.DevUse.UsedBy)
			return
		}
	}

	// checking for cstor and zfs localPV
	// we start with the assumption that device has a zfs file system
	hasZFS := true
	devPath := blockDevice.DevPath

	// if device is not a partition, check whether the zfs partitions are available on the disk
	if blockDevice.DeviceAttributes.DeviceType != libudevwrapper.UDEV_PARTITION {
		path, ok := getBlockDeviceZFSPartition(*blockDevice)
		if ok {
			devPath = path
		} else {
			hasZFS = false
			klog.V(4).Infof("device: %s is not having any zfs partitions", blockDevice.DevPath)
		}
	}

	if hasZFS {
		samplingProbe := newSamplingProbe(devPath)

		// check for ZFS file system
		fstype := samplingProbe.BlkidIdentifier.GetOnDiskFileSystem()

		if fstype == zfsFileSystemLabel {
			blockDevice.DevUse.InUse = true

			// the disk can either be in use by cstor or zfs local PV
			ok, err := isBlockDeviceInUse(blockDevice.DevPath)

			if err != nil {
				klog.Errorf("error checking block device: %s: %v", blockDevice.DevPath, err)
			}
			if ok {
				blockDevice.DevUse.UsedBy = blockdevice.ZFSLocalPV
			} else {
				blockDevice.DevUse.UsedBy = blockdevice.CStor
			}
			klog.V(4).Infof("device: %s Used by: %s filled by sampling probe", blockDevice.DevPath, blockDevice.DevUse.UsedBy)
			return
		}
	}

	// create a device identifier for reading the spdk super block from the disk
	spdkIdentifier := &spdk.DeviceIdentifier{
		DevPath: blockDevice.DevPath,
	}

	signature, err := spdkIdentifier.GetSPDKSuperBlockSignature()
	if err != nil {
		klog.Errorf("error reading spdk signature from device: %s, %v", blockDevice.DevPath, err)
	}
	if spdk.IsSPDKSignatureExist(signature) {
		blockDevice.DevUse.InUse = true
		blockDevice.DevUse.UsedBy = blockdevice.Mayastor
		return
	}

	// TODO jiva disk detection
}

// getBlockDeviceZFSPartition is used to get the zfs partition if it exist in a
// given BD
func getBlockDeviceZFSPartition(bd blockdevice.BlockDevice) (string, bool) {
	zfsDataPartitionNumber := "1"
	zfsMetaPartitionNumber := "9"

	// to handle cases of devices with nvme drives, the device name will be
	// nvme0n1
	if util.IsMatchRegex(".+[0-9]+$", bd.DevPath) {
		zfsDataPartitionNumber = "p" + zfsDataPartitionNumber
		zfsMetaPartitionNumber = "p" + zfsMetaPartitionNumber
	}

	dataPartition := bd.DevPath + zfsDataPartitionNumber
	metaPartition := bd.DevPath + zfsMetaPartitionNumber

	// check if device has the 2 partitions
	if bd.DependentDevices.Partitions[0] == dataPartition &&
		bd.DependentDevices.Partitions[1] == metaPartition {
		return dataPartition, true
	}
	return "", false
}

// isBlockDeviceInUse tries to open the device exclusively to check if the device is
// being held by some process. eg: If kernel zfs uses the disk, the open will fail
func isBlockDeviceInUse(path string) (bool, error) {
	_, err := os.OpenFile(path, os.O_EXCL, 0444)

	if errors.Is(err, syscall.EBUSY) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
