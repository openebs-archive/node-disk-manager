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

package udev

import (
	"fmt"
	"math/rand" // nosec G404
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
)

// Disk and system attributes corresponding to backing image and
// the loop device
const (
	imageDirectory = "/tmp"
	syspath        = "/sys/class/block"
)

// Disk has the attributes of a virtual disk which is emulated for integration
// testing.
type Disk struct {
	// Size in bytes
	Size int64
	// The backing image name
	// eg: /tmp/fake123
	imageName string
	// the disk name
	// eg: /dev/loop9002
	Name string
	// mount point if any
	MountPoints []string
}

// NewDisk creates a Disk struct, with a specified size. Also the
// random disk image name is generated. The actual image is generated only when
// we try to attach the disk
func NewDisk(size int64) Disk {
	disk := Disk{
		Size:      size,
		imageName: generateDiskImageName(),
		Name:      "",
	}
	return disk
}

func (disk *Disk) createDiskImage() error {
	// no of blocks
	/*count := disk.Size / blockSize
	createImageCommand := "dd if=/dev/zero of=" + disk.imageName + " bs=" + strconv.Itoa(blockSize) + " count=" + strconv.Itoa(int(count))
	err := utils.RunCommand(createImageCommand)*/
	f, err := os.Create(disk.imageName)
	if err != nil {
		return fmt.Errorf("error creating disk image. Error : %v", err)
	}
	err = f.Truncate(disk.Size)
	if err != nil {
		return fmt.Errorf("error truncating disk image. Error : %v", err)
	}

	return nil
}

func (disk *Disk) createLoopDevice() error {
	var err error
	if _, err = os.Stat(disk.imageName); err != nil {
		err = disk.createDiskImage()
		if err != nil {
			return err
		}
	}

	deviceName := getLoopDevName()
	devicePath := "/dev/" + deviceName
	// create the loop device using losetup
	createLoopDeviceCommand := "losetup " + devicePath + " " + disk.imageName
	err = utils.RunCommandWithSudo(createLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("error creating loop device. Error : %v", err)
	}
	disk.Name = devicePath
	return nil
}

// DetachAndDeleteDisk triggers a udev remove event. It detaches the loop device from the backing
// image. Also deletes the backing image and block device file in /dev
func (disk *Disk) DetachAndDeleteDisk() error {
	if disk.Name == "" {
		return fmt.Errorf("no such disk present for deletion")
	}
	detachLoopCommand := "losetup -d " + disk.Name
	err := utils.RunCommandWithSudo(detachLoopCommand)
	if err != nil {
		return fmt.Errorf("cannot detach loop device. Error : %v", err)
	}
	err = TriggerEvent(UdevEventRemove, syspath, disk.Name)
	if err != nil {
		return fmt.Errorf("could not trigger device remove event. Error : %v", err)
	}
	deleteBackingImageCommand := "rm " + disk.imageName
	err = utils.RunCommandWithSudo(deleteBackingImageCommand)
	if err != nil {
		return fmt.Errorf("could not delete backing disk image. Error : %v", err)
	}
	deleteLoopDeviceCommand := "rm " + disk.Name
	err = utils.RunCommandWithSudo(deleteLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("could not delete loop device. Error : %v", err)
	}
	return nil
}

// Generates a random image name for the backing file.
// the file name will be of the format fakeXXX, where X=[0-9]
func generateDiskImageName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := 100 + rand.Intn(899)
	imageName := "fake" + strconv.Itoa(randomNumber)
	return imageDirectory + "/" + imageName
}

// Generates a random loop device name. The name will be of the
// format loop9XXX, where X=[0-9]. The 4 digit numbering is chosen so that
// we get enough disks to be randomly generated and also it does not clash
// with the existing loop devices present in some systems.
func getLoopDevName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := 9000 + rand.Intn(999)
	diskName := "loop" + strconv.Itoa(randomNumber)
	return diskName
}

// AttachDisk triggers a udev add event for the disk. If the disk is not present, the loop
// device is created and event is triggered
func (disk *Disk) AttachDisk() error {
	if disk.Name == "" {
		if err := disk.createLoopDevice(); err != nil {
			return err
		}
	}
	return TriggerEvent(UdevEventAdd, syspath, disk.Name)
}

// DetachDisk triggers a udev remove event for the disk
func (disk *Disk) DetachDisk() error {
	return TriggerEvent(UdevEventRemove, syspath, disk.Name)
}

func (di *Disk) CreateFileSystem() error {
	var err error
	if _, err = os.Stat(di.imageName); err != nil {
		err = di.createDiskImage()
		if err != nil {
			return err
		}
	}

	d, err := diskfs.Open(di.imageName)
	defer d.File.Close()
	if err != nil {
		return err
	}

	_, err = d.CreateFilesystem(disk.FilesystemSpec{
		Partition: 0,
		FSType:    filesystem.TypeFat32,
	})

	if err != nil {
		return fmt.Errorf("failed to create file system: %v", err)
	}
	return nil
}

func (disk *Disk) Mount(path string) error {
	path = filepath.Clean(path)
	err := syscall.Mount(disk.Name, path, "vfat", 0, "")
	if err != nil {
		return err
	}

	disk.MountPoints = append(disk.MountPoints, path)
	return nil
}

func (disk *Disk) Unmount() error {
	var lastErr error = nil
	for _, mp := range disk.MountPoints {
		err := syscall.Unmount(mp, 0)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (disk *Disk) Resize(size int64) error {
	units := [4]string{"", "K", "M", "G"}
	currUnit := -1
	var maxExp int64 = 1
	for maxExp <= size {
		maxExp *= 1024
		currUnit++
	}
	maxExp /= 1024
	size /= maxExp
	err := utils.RunCommand(fmt.Sprintf("dd if=/dev/zero of=%s bs=1%s count=%d",
		disk.imageName, units[currUnit], size))
	if err != nil {
		return err
	}
	return utils.RunCommandWithSudo("losetup -c " + disk.Name)
}
