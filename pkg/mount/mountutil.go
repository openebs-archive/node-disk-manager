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

package mount

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type MountUtil struct {
	filePath string
	devPath  string
}

// newMountUtil returns MountUtil struct for given mounts file path and mount point
func newMountUtil(filePath, devPath string) MountUtil {
	MountUtil := MountUtil{
		filePath: filePath,
		devPath:  devPath,
	}
	return MountUtil
}

// getPartitionName read mounts file and returns partition name which is mounted on mount point
func (m MountUtil) getMountAttr() (MountAttr, error) {
	mountAttr := MountAttr{}
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
		// 1st entry is partition or file system 2nd is mount point
		if parts := strings.Split(line, " "); parts[0] == m.devPath {
			mountAttr.MountPoint = parts[1]
			mountAttr.FileSystem = parts[2]
			return mountAttr, nil
		}
	}
	return mountAttr, fmt.Errorf("could not get mount attributes, %s not mounted", m.devPath)
}
