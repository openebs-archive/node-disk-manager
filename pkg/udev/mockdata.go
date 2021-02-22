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

// +build linux,cgo

package udev

/*
  #cgo LDFLAGS: -ludev
  #include <libudev.h>
*/
import "C"
import (
	"bufio"
	bd "github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/mount"
	"github.com/openebs/node-disk-manager/pkg/sysfs"
	"io/ioutil"
	"os"
	"strings"
)

// MockOsDiskDetails struct contain different attribute of os disk.
type MockOsDiskDetails struct {
	OsDiskName     string
	DevType        string
	DevNode        string
	Size           string
	SysPath        string
	Model          string
	Serial         string
	Vendor         string
	Wwn            string
	Uid            string
	FileSystem     string
	Mountpoint     string
	PartTableType  string
	PartTableUUID  string
	IdType         string
	ByIdDevLinks   []string
	ByPathDevLinks []string
	Dependents     bd.DependentBlockDevices
}

// mockDataStructUdev returns C udev struct for unit test.
func mockDataStructUdev() *C.struct_udev {
	return C.udev_new()
}

// MockDiskDetails returns os disk details which is used in unit test.
func MockDiskDetails() (MockOsDiskDetails, error) {
	diskDetails := MockOsDiskDetails{}
	osDiskName, osFilesystem, err := OsDiskName()
	if err != nil {
		return diskDetails, err
	}
	sysPath, err := getSyspathOfOsDisk(osDiskName)
	if err != nil {
		return diskDetails, err
	}
	size, err := getOsDiskSize(osDiskName)
	if err != nil {
		return diskDetails, err
	}
	device := getOsDiskUdevDevice(sysPath)
	if device == nil {
		return diskDetails, err
	}
	defer device.UdevDeviceUnref()
	diskDetails.OsDiskName = osDiskName
	diskDetails.DevType = "disk"
	diskDetails.DevNode = "/dev/" + osDiskName
	diskDetails.Size = size
	diskDetails.SysPath = sysPath
	diskDetails.Model = device.GetPropertyValue(UDEV_MODEL)
	diskDetails.Serial = device.GetPropertyValue(UDEV_SERIAL)
	diskDetails.Vendor = device.GetPropertyValue(UDEV_VENDOR)
	diskDetails.Wwn = device.GetPropertyValue(UDEV_WWN)
	diskDetails.FileSystem = osFilesystem
	diskDetails.PartTableType = device.GetPropertyValue(UDEV_PARTITION_TABLE_TYPE)
	diskDetails.PartTableUUID = device.GetPropertyValue(UDEV_PARTITION_TABLE_UUID)
	diskDetails.IdType = device.GetPropertyValue(UDEV_TYPE)
	dev, err := sysfs.NewSysFsDeviceFromDevPath(diskDetails.DevNode)
	if err != nil {
		return diskDetails, err
	}
	diskDetails.Dependents, err = dev.GetDependents()
	if err != nil {
		return diskDetails, err
	}
	diskDetails.Mountpoint = "/" // always take the disk mounted at /
	devLinks := device.GetDevLinks()
	diskDetails.ByIdDevLinks = devLinks[BY_ID_LINK]
	diskDetails.ByPathDevLinks = devLinks[BY_PATH_LINK]
	return diskDetails, nil
}

// OsDiskName returns os disk name given by kernel
func OsDiskName() (string, string, error) {
	var osPartPath, osFileSystem string
	// Read /proc/self/mounts file to get which partition is mounted on / path.
	file, err := os.Open("/proc/self/mounts")
	if err != nil {
		return osPartPath, osFileSystem, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[1] == "/" {
			// Get dev path of the partition which is mounted on / path
			osPartPath = parts[0]
			osFileSystem = parts[2]
			break
		}
	}

	mountPointUtil := mount.NewMountUtil("/proc/self/mounts", "", "/")
	disk, err := mountPointUtil.GetDiskPath()
	if err != nil {
		return osPartPath, osFileSystem, err
	}
	disk = strings.Replace(disk, "/dev/", "", 1)
	return disk, osFileSystem, nil
}

// getSyspathOfOsDisk returns syspath of os disk in success
func getSyspathOfOsDisk(osDiskName string) (string, error) {
	data, err := ioutil.ReadFile("/sys/class/block/" + osDiskName + "/dev")
	if err != nil {
		return "", err
	}
	return "/sys/dev/block/" + strings.TrimSpace(string(data)), nil
}

// getOsDiskSize returns size of os disk in success
func getOsDiskSize(osDiskName string) (string, error) {
	sizeByte, err := ioutil.ReadFile("/sys/class/block/" + osDiskName + "/size")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(sizeByte)), nil
}

// getOsDiskUdevDevice UdevDevice struct of Os disk
func getOsDiskUdevDevice(sysPath string) *UdevDevice {
	udev, err := NewUdev()
	if err != nil {
		return nil
	}
	device, err := udev.NewDeviceFromSysPath(sysPath)
	if err != nil {
		return nil
	}
	return device
}
