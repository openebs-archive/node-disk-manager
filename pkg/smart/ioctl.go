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

// See Linux man-pages http://man7.org/linux/man-pages/man2/capset.2.html

package smart

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	linuxCapabilityVersion3 = 0x20080522 // linux capability version 3
	capSysRawIO             = 1 << 17    // input/output operations capabilities
	capSysAdmin             = 1 << 21    // admin capabilities
)

// Header struct to be used while sending capget syscall
type userCapHeader struct {
	version uint32 // version of the capget syscall
	pid     int    // process id of thread whose capabilities needs to be checked
}

// The effective, permitted, and inheritable fields are bit masks of the
// capabilities defined in capabilities(7).
// userCapData struct needs to be passed as an argument to the capget system call to check
// the capabilities of a thread.
type userCapData struct {
	effective   uint32
	permitted   uint32
	inheritable uint32
}

// userCapsV3 is used to form the struct to send a capability version 3 capget syscall
type userCapsV3 struct {
	hdr  userCapHeader
	data [2]userCapData
}

// CheckBinaryPerm invokes the linux CAPGET syscall which checks for necessary capabilities required for a binary to access a device.
// Note that this depends on the binary having the capabilities set
// (i.e., via the `setcap` utility), and on VFS support i.e. with VFS support,
// for capset() calls the only permitted values for userCapHeader->pid are 0
// Note : If the binary is executed as root, it automatically has all capabilities set.
func CheckBinaryPerm() error {
	userCaps := new(userCapsV3)
	userCaps.hdr.version = linuxCapabilityVersion3

	// Performs a raw system call to check the capabilities set, returns an error no
	// other than 0 if capget system call fails otherwise 0
	// An Errno is an unsigned number describing an error condition.
	// It implements the error interface. The zero Errno is by convention
	// a non-error else error.
	_, _, errno := unix.RawSyscall(unix.SYS_CAPGET, uintptr(unsafe.Pointer(&userCaps.hdr)), uintptr(unsafe.Pointer(&userCaps.data)), 0)
	if errno != 0 {
		return fmt.Errorf("linux capget syscall has failed: %+v", errno)
	}
	// Check if capSysRawIO (Perform I/O port operations,perform various SCSI device commands,etc)
	// and capSysAdmin (Perform a range of system administration operations,
	// perform various privileged block-device and filesystem ioctl operations) are in effect
	if (userCaps.data[0].effective&capSysRawIO == 0) && (userCaps.data[0].effective&capSysAdmin == 0) {
		return fmt.Errorf("capSysRawIO and capSysAdmin are not in effect, device access will fail")
	}
	return nil
}

// Ioctl function executes an ioctl command on the specified file descriptor
// ioctl (an abbreviation of input/output control) is a system call for device-specific
// input/output operations and other operations which cannot be expressed by regular system calls.
// It takes a parameter specifying a request code; the effect of a call depends completely on the request code
func Ioctl(fd, cmd, ptr uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, cmd, ptr)
	// An Errno is an unsigned number describing an error condition.
	// It implements the error interface. The zero Errno is by convention
	// a non-error else error.
	if errno != 0 {
		return errno
	}
	return nil
}
