/*
Copyright 2018 The OpenEBS Authors.

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

package util

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// MountUtil contains mount point and mount file attributes
// It helps to find which partition is mounted on given mount point
type MountUtil struct {
	FilePath   string // FilePath is the path of mounts file like -/proc/self/mounts
	MountPoint string // MountPoint is mount points like - / /var ...
}

// NewMountUtil returns MountUtil struct for given mounts file path and mount point
func NewMountUtil(filePath, mountPoint string) MountUtil {
	MountUtil := MountUtil{
		FilePath:   filePath,
		MountPoint: mountPoint,
	}
	return MountUtil
}

// GetDiskPath returns os disk devpath
func (m MountUtil) GetDiskPath() (string, error) {
	partition, err := m.getPartitionName()
	if err != nil {
		return "", err
	}
	devPath, err := getDiskDevPath(partition)
	if err != nil {
		return "", err
	}
	_, err = filepath.EvalSymlinks(devPath)
	if err != nil {
		return "", err
	}
	return devPath, err
}

// getPartitionName read mounts file and returns partition name which is mounted on mount point
func (m MountUtil) getPartitionName() (string, error) {
	// Read file from filepath and get which partition is mounted on given mount point
	file, err := os.Open(m.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return "", err
	}
	for scanner.Scan() {
		line := scanner.Text()
		/*
			read each line of given file in below format -
			/dev/sda4 / ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
			/dev/sda4 /var/lib/docker/aufs ext4 rw,relatime,errors=remount-ro,data=ordered 0 0
			1st entry is partition or file system 2nd is mount point
		*/
		if parts := strings.Split(line, " "); parts[1] == m.MountPoint &&
			// only if /dev prefix, otherwise it can be filesystem
			strings.HasPrefix(parts[0], "/dev") {
			// /dev/ by default added with partition name we want to get only sda1 / sdc2 ..
			return strings.Replace(parts[0], "/dev/", "", 1), nil
		}
	}
	return "", errors.New("error while geting os partition name")
}

/*
	getDiskSysPath takes disk/partition name as input (sda, sda1, sdb, sdb2 ...) and
	returns syspath of that disk from which we can generate ndm given uuid of that disk.
*/
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
