package controller

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiskToDeviceUUID(t *testing.T) {
	ctrl := Controller{}
	diskUID := "disk-adb1b23f7ba395988d029d78ef7bda58"
	blockDeviceUID := "blockdevice-adb1b23f7ba395988d029d78ef7bda58"
	assert.Equal(t, blockDeviceUID, ctrl.DiskToDeviceUUID(diskUID))
}
