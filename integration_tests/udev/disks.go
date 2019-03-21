package udev

import (
	"fmt"
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"math/rand"
	"strconv"
	"time"
)

const (
	blockSize      = 4096
	imageDirectory = "/tmp"
	syspath        = "/sys/class/block"
)

type Disk struct {
	// Size in bytes
	Size uint64
	// The backing image name
	// eg: /tmp/fake123
	imageName string
	// the disk name
	// eg: /dev/loop9002
	Name string
}

func NewDisk(size uint64) Disk {
	disk := Disk{
		Size:      size,
		imageName: generateDiskImageName(),
		Name:      "",
	}
	return disk
}

// Create a fake disk. The function creates a file backed loop
// device and udev add event will also be triggered
func (disk *Disk) createAndAttachDisk() error {
	// no of blocks
	count := disk.Size / blockSize
	createImageCommand := "dd if=/dev/zero of=" + disk.imageName + " bs=" + strconv.Itoa(blockSize) + " count=" + strconv.Itoa(int(count))
	err := utils.RunCommand(createImageCommand)
	if err != nil {
		return fmt.Errorf("error creating disk image. Error : %v", err)
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

// Triggers a udev remove event. Detaches the loop device from the backing
// image. Also deletes the backing image and block device file in /dev
func (disk *Disk) DetachAndDeleteDisk() error {
	if disk.Name == "" {
		return fmt.Errorf("no such present for deletion")
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

// Triggers a udev add event for the disk. If the disk is not present, the loop
// device is created and event is triggered
func (disk *Disk) AttachDisk() error {
	if disk.Name == "" {
		return disk.createAndAttachDisk()
	}
	return TriggerEvent(UdevEventAdd, syspath, disk.Name)
}

// Detach the disk, a udev remove event is triggered for the disk
func (disk *Disk) DetachDisk() error {
	return TriggerEvent(UdevEventRemove, syspath, disk.Name)
}
