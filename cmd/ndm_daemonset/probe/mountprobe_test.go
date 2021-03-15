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

package probe

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/mount"
)

const (
	mountsFilePath   = "/proc/1/mounts"
	sampleMountsFile = `
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
udev /dev devtmpfs rw,nosuid,noexec,relatime,size=4027936k,nr_inodes=1006984,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,nodev,noexec,relatime,size=811944k,mode=755 0 0
/dev/sda1 / ext4 rw,relatime,errors=remount-ro 0 0
none /sys/fs/bpf bpf rw,nosuid,nodev,noexec,relatime,mode=700 0 0
cgroup /sys/fs/cgroup/rdma cgroup rw,nosuid,nodev,noexec,relatime,rdma 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
/dev/loop4 /snap/core18/1988 squashfs ro,nodev,relatime 0 0
/dev/loop0 /snap/core/10823 squashfs ro,nodev,relatime 0 0
/dev/loop3 /snap/gnome-3-28-1804/145 squashfs ro,nodev,relatime 0 0
/dev/loop1 /snap/core/10859 squashfs ro,nodev,relatime 0 0
/dev/loop2 /snap/core18/1944 squashfs ro,nodev,relatime 0 0
/dev/sda4 /boot/efi vfat rw,relatime,fmask=0022,dmask=0022,codepage=437,iocharset=iso8859-1,shortname=mixed,errors=remount-ro 0 0
/dev/sda2 /home ext4 rw,relatime 0 0
`
)

func TestMountProbeFillBlockDeviceDetails(t *testing.T) {
	mp := &mountProbe{}

	t.Run("DevPath empty", func(t *testing.T) {
		bd := blockdevice.BlockDevice{}
		mp.FillBlockDeviceDetails(&bd)
		if len(bd.FSInfo.MountPoint) > 0 {
			t.Errorf("Expected mountpoints to be empty, found %v",
				bd.FSInfo.MountPoint)
		}
	})

	t.Run("Device not mounted", func(t *testing.T) {
		// Chroot into a tmp dir
		err := enterFakeRoot(t)
		if err != nil {
			t.Fatal(err)
		}
		defer exitFakeRoot()

		bd := blockdevice.BlockDevice{
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda5",
			},
		}
		mp.FillBlockDeviceDetails(&bd)
		if len(bd.FSInfo.MountPoint) > 0 {
			t.Errorf("Expected mountpoints to be empty, found %v",
				bd.FSInfo.MountPoint)
		}
	})

	t.Run("Device mounted", func(t *testing.T) {
		err := enterFakeRoot(t)
		if err != nil {
			t.Fatal(err)
		}
		defer exitFakeRoot()

		bd := blockdevice.BlockDevice{
			Identifier: blockdevice.Identifier{
				DevPath: "/dev/sda1",
			}}
		mp.FillBlockDeviceDetails(&bd)
		if len(bd.FSInfo.MountPoint) <= 0 || bd.FSInfo.MountPoint[0] != "/" {
			t.Errorf("Expected mountpoint to be %v, found %v",
				"/", bd.FSInfo.MountPoint)
		}
	})
}

// enterFakeRoot creates a fake root in a tmp folder and
// chroots to that folder.
func enterFakeRoot(t *testing.T) error {
	fakeRootPath := t.TempDir()
	realMountsPath := path.Join(fakeRootPath, mount.HostMountsFilePath[1:])

	t.Helper()

	// Create the mounts file at the expected path in the tmp dir
	err := os.MkdirAll(path.Dir(realMountsPath), 0755)
	if err != nil {
		return err
	}
	err = createMountsFile(realMountsPath)
	if err != nil {
		return err
	}

	// Change dir to current root to get chroot back later
	err = os.Chdir("/")
	if err != nil {
		return err
	}

	// Chroot into the tmp dir
	err = syscall.Chroot(fakeRootPath)
	if err != nil {
		return err
	}
	return nil
}

// exitFakeRoot exits from the chrooted environment
func exitFakeRoot() error {
	// We need to chroot to the original root. The current dir
	// is the actual root since we changed dir to / while
	// entering fake root
	if err := syscall.Chroot("."); err != nil {
		return err
	}
	return nil
}

func createMountsFile(dest string) error {
	return ioutil.WriteFile(dest, []byte(sampleMountsFile), 0444)
}
