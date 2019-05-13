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
	"os"
	"path/filepath"
	"strings"
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
	devPath, err := getDiskDevPath(mountAttr.DevPath)
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
		if mountAttr, ok := fn(line); ok {
			return mountAttr, nil
		}
	}
	return mountAttr, fmt.Errorf("could not get device mount attributes, Path/MountPoint not present in mounts file")
}

//	getDiskSysPath takes disk/partition name as input (sda, sda1, sdb, sdb2 ...) and
//	returns syspath of that disk from which we can generate ndm given uuid of that disk.
func getDiskDevPath(partition string) (string, error) {
	// dev path be like /dev/sda4 we need to remove /dev/ from this string to get sys block path.
	var diskName string
	softlink := "/sys/class/block/" + partition
	link, err := filepath.EvalSymlinks(softlink)
	if err != nil {
		return "", err
	}
	/*
		link looks - /sys/devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4
		parent disk is present after block then partition of that disk
	*/
	parts := strings.Split(link, "/")
	for i, part := range parts {
		if part == "block" {
			diskName = parts[i+1]
			break
		}
	}
	return "/dev/" + diskName, nil
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
		mountAttr.DevPath = strings.Replace(parts[0], "/dev/", "", 1)
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
		mountAttr.MountPoint = parts[1]
		mountAttr.FileSystem = parts[2]
		isValid = true
	}
	return mountAttr, isValid
}
