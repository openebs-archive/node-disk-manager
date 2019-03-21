package udev

import (
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"strings"
)

const (
	UdevEventAdd    = "add"
	UdevEventRemove = "remove"
	UdevEventChange = "change"
)

var execCommandWithPipe = utils.ExecCommandWithPipe

// Triggers the udev event for the device with syspath
func TriggerEvent(event, syspath, devicePath string) error {
	splitName := strings.Split(devicePath, "/")
	deviceName := splitName[len(splitName)-1]
	fileName := syspath + "/" + deviceName + "/uevent"
	command1 := "echo " + event
	command2 := "sudo tee " + fileName
	_, err := execCommandWithPipe(command1, command2)
	return err
}
