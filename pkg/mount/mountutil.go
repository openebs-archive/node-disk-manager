/*
Copyright 2019 The OpenEBS Authors.

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

package mount

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openebs/node-disk-manager/pkg/features"
)

var ErrCouldNotFindRootDevice = fmt.Errorf("could not find root device")

const (
	procCmdLine     = "/proc/cmdline"
	hostProcCmdLine = "/host" + procCmdLine
)

// DiskMountUtil contains the mountfile path, devpath/mountpoint which can be used to
// detect partition of a mountpoint or mountpoint of a partition.
type DiskMountUtil struct {
	filePath   string
	devPath    string
	mountPoint string
}

type getMountData func(string) (DeviceMountAttr, bool)

// NewMountUtil returns DiskMountUtil struct for given mounts file path and mount point
func NewMountUtil(filePath, devPath, mountPoint string) DiskMountUtil {
	MountUtil := DiskMountUtil{
		filePath:   filePath,
		devPath:    devPath,
		mountPoint: mountPoint,
	}
	return MountUtil
}

// GetDiskPath returns os disk devpath
func (m DiskMountUtil) GetDiskPath() (string, error) {
	mountAttr, err := m.getDeviceMountAttr(m.getPartitionName)
	if err != nil {
		return "", err
	}
	devPath, err := getPartitionDevPath(mountAttr.DevPath)
	if err != nil {
		return "", err
	}
	_, err = filepath.EvalSymlinks(devPath)
	if err != nil {
		return "", err
	}
	return devPath, err
}

// getDeviceMountAttr read mounts file and returns device mount attributes, which includes partition name,
// mountpoint and filesystem
func (m DiskMountUtil) getDeviceMountAttr(fn getMountData) (DeviceMountAttr, error) {
	found := false
	mountAttr := DeviceMountAttr{}
	// Read file from filepath and get which partition is mounted on given mount point
	file, err := os.Open(m.filePath)
	if err != nil {
		return mountAttr, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return mountAttr, err
	}
	for scanner.Scan() {
		line := scanner.Text()

		/*
			read each line of given file in below format -
			/dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
			/dev/sda4 /var/lib/docker/aufs ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
		*/

		// we are interested only in lines that start with /dev
		if !strings.HasPrefix(line, "/dev") {
			continue
		}
		if lineMountAttr, ok := fn(line); ok {
			found = true
			mergeDeviceMountAttrs(&mountAttr, &lineMountAttr)
		}
	}
	if found {
		return mountAttr, nil
	}
	return mountAttr, fmt.Errorf("could not get device mount attributes, Path/MountPoint not present in mounts file")
}

//	getPartitionDevPath takes disk/partition name as input (sda, sda1, sdb, sdb2 ...) and
//	returns dev path of that disk/partition (/dev/sda1,/dev/sda)
//
// NOTE: if the feature gate to use OS disk is enabled, the dev path of disk /partition is returned,
// eg: sda1, sda2, root on sda5 returns /dev/sda1, /dev/sda2, /dev/sda5 respectively
// else, the devpath of the parent disk will be returned
// eg: sda1, root on sda5 returns /dev/sda, /dev/sda respectively
func getPartitionDevPath(partition string) (string, error) {
	softlink, err := getSoftLinkForPartition(partition)
	if err != nil {
		return "", err
	}

	link, err := filepath.EvalSymlinks(softlink)
	if err != nil {
		return "", err
	}

	var disk string
	var ok bool
	if features.FeatureGates.IsEnabled(features.UseOSDisk) {
		// the last part will be used instead of the parent disk
		// eg: /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4 is the link
		// and sda4 will be the device.
		split := strings.Split(link, "/")
		disk = split[len(split)-1]
	} else {
		disk, ok = getParentBlockDevice(link)
		if !ok {
			return "", fmt.Errorf("could not find parent device for %s", link)
		}
	}

	return "/dev/" + disk, nil
}

//	getSoftLinkForPartition returns path to /sys/class/block/$partition
//	if the path does not exist and the partition is "root"
//	then the root partition is detected from /proc/cmdline
func getSoftLinkForPartition(partition string) (string, error) {
	softlink := getLinkForPartition(partition)

	if !fileExists(softlink) && partition == "root" {
		partition, err := getRootPartition()
		if err != nil {
			return "", err
		}
		softlink = getLinkForPartition(partition)
	}
	return softlink, nil
}

//	getLinkForPartition returns path to sys block path
func getLinkForPartition(partition string) string {
	// dev path be like /dev/sda4 we need to remove /dev/ from this string to get sys block path.
	return "/sys/class/block/" + partition
}

//	getRootPartition resolves link /dev/root using /proc/cmdline
func getRootPartition() (string, error) {
	file, err := os.Open(getCmdlineFile())
	if err != nil {
		return "", err
	}
	defer file.Close()

	path, err := parseRootDeviceLink(file)
	if err != nil {
		return "", err
	}

	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	return getDeviceName(link), nil
}

func parseRootDeviceLink(file io.Reader) (string, error) {
	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return "", err
	}

	rootPrefix := "root="
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		args := strings.Split(line, " ")

		// looking for root device identification
		// ... root=UUID=d41162ba-25e4-4c44-8793-2abef96d27e9 ...
		for _, arg := range args {
			if !strings.HasPrefix(arg, rootPrefix) {
				continue
			}

			rootSpec := strings.Split(arg[len(rootPrefix):], "=")

			// if the expected format is not present, then we skip getting the root partition
			if len(rootSpec) < 2 {
				if strings.HasPrefix(rootSpec[0], "/dev") {
					return rootSpec[0], nil
				}
				return "", ErrCouldNotFindRootDevice
			}

			identifierType := strings.ToLower(rootSpec[0])
			identifier := rootSpec[1]

			return fmt.Sprintf("/dev/disk/by-%s/%s", identifierType, identifier), nil
		}
	}

	return "", ErrCouldNotFindRootDevice
}

// getParentBlockDevice returns the parent blockdevice of a given blockdevice by parsing the syspath
//
// syspath looks like - /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4
// parent disk is present after block then partition of that disk
//
// for blockdevices that belong to the nvme subsystem, the symlink has a different format,
// /sys/devices/pci0000:00/0000:00:0e.0/nvme/nvme0/nvme0n1/nvme0n1p1. So we search for the nvme subsystem
// instead of `block`. The blockdevice will be available after the NVMe instance, nvme/instance/namespace.
// The namespace will be the blockdevice.
func getParentBlockDevice(sysPath string) (string, bool) {
	blockSubsystem := "block"
	nvmeSubsystem := "nvme"
	parts := strings.Split(sysPath, "/")

	// checking for block subsystem, return the next part after subsystem only
	// if the length is greater. This check is to avoid an index out of range panic.
	for i, part := range parts {
		if part == blockSubsystem &&
			len(parts)-1 >= i+1 {
			return parts[i+1], true
		}
	}

	// checking for nvme subsystem, return the 2nd item in hierarchy, which will be the
	// nvme namespace. Length checking is to avoid index out of range in case of malformed
	// links (extremely rare case)
	for i, part := range parts {
		if part == nvmeSubsystem &&
			len(parts)-1 >= i+2 {
			return parts[i+2], true
		}
	}
	return "", false
}

// getPartitionName gets the partition name from the mountpoint. Each line of a mounts file
// is passed to the function, which is parsed and partition name is obtained
// A mountLine contains data in the order:
// 		device  mountpoint  filesystem  mountoptions
//		eg: /dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
func (m *DiskMountUtil) getPartitionName(mountLine string) (DeviceMountAttr, bool) {
	mountAttr := DeviceMountAttr{}
	isValid := false
	if len(mountLine) == 0 {
		return mountAttr, isValid
	}
	// mountoptions are ignored. device-path and mountpoint is used
	if parts := strings.Split(mountLine, " "); parts[1] == m.mountPoint {
		mountAttr.DevPath = getDeviceName(parts[0])
		isValid = true
	}
	return mountAttr, isValid
}

// getMountName gets the mountpoint, filesystem etc from the partition name. Each line of a mounts
// file is passed to the function, which is parsed to get the information
// A mountLine contains data in the order:
// 		device  mountpoint  filesystem  mountoptions
//		eg: /dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
func (m *DiskMountUtil) getMountName(mountLine string) (DeviceMountAttr, bool) {
	mountAttr := DeviceMountAttr{}
	isValid := false
	if len(mountLine) == 0 {
		return mountAttr, isValid
	}
	// mountoptions are ignored. devicepath, mountpoint and filesystem is used
	if parts := strings.Split(mountLine, " "); parts[0] == m.devPath {
		mountAttr.MountPoint = []string{parts[1]}
		mountAttr.FileSystem = parts[2]
		isValid = true
	}
	return mountAttr, isValid
}

func getCmdlineFile() string {
	if fileExists(hostProcCmdLine) {
		return hostProcCmdLine
	}
	return procCmdLine
}

// getDeviceName gets the blockdevice special file name.
// eg: sda, sdb
// if a mapper device is specified the symlink will be evaluated and the
// dm-X name will be returned
func getDeviceName(devPath string) string {
	var err error
	var deviceName string

	deviceName = devPath
	// if the device is a dm device
	if strings.HasPrefix(devPath, "/dev/mapper") {
		deviceName, err = filepath.EvalSymlinks(devPath)
		if err != nil {
			return ""
		}
	}
	return strings.Replace(deviceName, "/dev/", "", 1)
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// mergeDeviceMountAttrs merges the second mountattr into the first. The merge is
// performed as follows:
// 1. If the DevPath of the first mountattr is empty, then it is set to the DevPath of
// the second mountattr
// 2. If the FileSystem of the first mountattr is empty, it is set to the FileSystem of
// the second mountattr provided that the DevPaths of both the mountattrs match
// 3. The MountPoint(s) of the second mountattr are appended to first's only if the
// DevPath and the FileSystem of both the mountattrs match and the FileSystem of first
// mountattr is non-empty (!= "")
func mergeDeviceMountAttrs(ma *DeviceMountAttr, mb *DeviceMountAttr) {
	if ma.DevPath == "" {
		ma.DevPath = mb.DevPath
	}
	if ma.DevPath != mb.DevPath {
		return
	}
	if ma.FileSystem == "" {
		ma.FileSystem = mb.FileSystem
	}
	if ma.FileSystem != "" && ma.FileSystem == mb.FileSystem {
		ma.MountPoint = append(ma.MountPoint, mb.MountPoint...)
	}
}
