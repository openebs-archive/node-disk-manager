// +build linux,cgo

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

package udev

/*
  #cgo LDFLAGS: -ludev
  #include <libudev.h>
*/
import "C"
import (
	"bufio"
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
	ByIdDevLinks   []string
	ByPathDevLinks []string
}

// mockDataStructUdev returns C udev struct for unit test.
func mockDataStructUdev() *C.struct_udev {
	return C.udev_new()
}

// MockDiskDetails returns os disk details which is used in unit test.
func MockDiskDetails() (MockOsDiskDetails, error) {
	diskDetails := MockOsDiskDetails{}
	osDiskName, err := OsDiskName()
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
	devLinks := device.GetDevLinks()
	diskDetails.ByIdDevLinks = devLinks[BY_ID_LINK]
	diskDetails.ByPathDevLinks = devLinks[BY_PATH_LINK]
	return diskDetails, nil
}

// OsDiskName returns os disk name given by kernel
func OsDiskName() (string, error) {
	var osPartPath string
	// Read /proc/self/mounts file to get which partition is mounted on / path.
	file, err := os.Open("/proc/self/mounts")
	if err != nil {
		return osPartPath, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[1] == "/" {
			// Get dev path of the partition which is mounted on / path
			osPartPath = parts[0]
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return osPartPath, err
	}
	// dev path be like /dev/sda4 we need to remove /dev/ from this string to get sys block path.
	osPartPath = strings.Replace(osPartPath, "/dev/", "", 1)
	softlink := "/sys/class/block/" + osPartPath
	link, err := os.Readlink(softlink)
	if err != nil {
		return osPartPath, err
	}
	parts := strings.Split(link, "/")
	if parts[len(parts)-2] != "block" {
		// If the link path is not for parent disk we need to remove partition name from link.
		// link looks like ../../devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda/sda4
		// or ../../devices/pci0000:00/0000:00:1f.2/ata1/host0/target0:0:0/0:0:0:0/block/sda
		// where sda is the parent disk if sda1 / sda2 .. partition present in link
		// we need to remove /sda4 from the link to get parent disk
		link = strings.Replace(link, "/"+osPartPath, "", 1)
	}
	parts = strings.Split(link, "/")
	osDiskName := parts[len(parts)-1]
	return osDiskName, nil
}

// getSyspathOfOsDisk returns syspath of os disk in success
func getSyspathOfOsDisk(osDiskName string) (string, error) {
	data, err := ioutil.ReadFile("/sys/block/" + osDiskName + "/dev")
	if err != nil {
		return "", err
	}
	return "/sys/dev/block/" + strings.TrimSpace(string(data)), nil
}

// getOsDiskSize returns size of os disk in success
func getOsDiskSize(osDiskName string) (string, error) {
	sizeByte, err := ioutil.ReadFile("/sys/block/" + osDiskName + "/size")
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
