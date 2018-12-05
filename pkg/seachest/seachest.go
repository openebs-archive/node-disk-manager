// +build linux,cgo

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

package seachest

/*
#cgo LDFLAGS: -lopensea-operations -lopensea-transport -lopensea-common
#cgo CFLAGS: -I../../../openSeaChest/include -I../../../openSeaChest/opensea-common/include -I../../../openSeaChest/opensea-operations/include -I../../../openSeaChest/opensea-transport/include
#include "common.h"
#include "openseachest_util_options.h"
#include "common_public.h"
#include "ata_helper.h"
#include "ata_helper_func.h"
#include "scsi_helper.h"
#include "scsi_helper_func.h"
#include "nvme_helper.h"
#include "nvme_helper_func.h"
#include "cmds.h"
#include "drive_info.h"
#include <libudev.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
	"github.com/golang/glog"
)

// Seachest errors are converted to string using this function
func SeachestErrors(err int) string {
	seachestErrorSting := []string {
		"Success",
		"Failed",
		"Not Supported",
		"Cmd Failure",
		"In-progress",
		"Aborted",
		"Bad Parameter",
		"Memory Allocation Failed",
		"Cmd Passthrough Failed",
		"Library Mistmatch",
		"Device Forzen",
		"Permission Denied",
		"File Open Error",
		"Incomplete RFTRS",
		"Cmd Time-out",
		"Warning - Not all device enumerated",
		"Invalid Checksum",
		"Cmd not Available",
		"Cmd Blocked",
		"Cmd Interrupted",
		"Unknown",
         }
         return seachestErrorSting[err]
}

// Identifier (devPath such as /dev/sda,etc) is an identifier for seachest probe
type Identifier struct {
	DevPath string
}

func (I *Identifier) SeachestBasicDiskInfo() (*C.driveInformationSAS_SATA, int) {

	var device C.tDevice
	var Drive C.driveInformationSAS_SATA
	str := C.CString(I.DevPath)
	defer C.free(unsafe.Pointer(str))

	err := int(C.get_Device(str, &device))
	if err != 0 {
		glog.Error("Unable to get device info for device:%s with error:%s", I.DevPath, SeachestErrors(err))
		return nil, err
	}

	err = int(C.get_SCSI_Drive_Information(&device, &Drive))
	if err != 0 {
		glog.Error("Unable to get derive info for device:%s with error:%s", I.DevPath, SeachestErrors(err))
		return nil, err
	}

	//Added for cross verification, can be removed later on.
	C.print_SAS_Sata_Device_Information(&Drive)

	return &Drive, err
}

func (I *Identifier) GetHostName(driveInfo *C.driveInformationSAS_SATA) string {
	return ""
}

func (I *Identifier) GetModelNumber(driveInfo *C.driveInformationSAS_SATA) string {
	var ptr *C.char
	ptr = &driveInfo.modelNumber[0]
	str := C.GoString(ptr)
	return str
}

func (I *Identifier) GetUuid(driveInfo *C.driveInformationSAS_SATA) string {
	myString := fmt.Sprintf("%v", driveInfo.worldWideName)
	return myString
}

func (I *Identifier) GetCapacity(driveInfo *C.driveInformationSAS_SATA) uint64 {
	var capacity C.uint64_t
	capacity = (C.uint64_t)(driveInfo.maxLBA * ((C.uint64_t)(driveInfo.logicalSectorSize)))
	return ((uint64)(capacity))
}

func (I *Identifier) GetSerialNumber(driveInfo *C.driveInformationSAS_SATA) string {
	var ptr *C.char
	ptr = &driveInfo.serialNumber[0]
	str := C.GoString(ptr)
	return str
}

func (I *Identifier) GetVendorID(driveInfo *C.driveInformationSAS_SATA) string {
	var ptr *C.char
	ptr = &driveInfo.vendorID[0]
	str := C.GoString(ptr)
	return str
}

func (I *Identifier) GetPath(driveInfo *C.driveInformationSAS_SATA) string {
	return I.DevPath
}

func (I *Identifier) GetFirmwareRevision(driveInfo *C.driveInformationSAS_SATA) string {
	var ptr *C.char
	ptr = &driveInfo.firmwareRevision[0]
	str := C.GoString(ptr)
	return str
}

func (I *Identifier) GetLogicalSectorSize(driveInfo *C.driveInformationSAS_SATA) uint32 {
	return ((uint32)(driveInfo.logicalSectorSize))
}

func (I *Identifier) IsRotational(driveInfo *C.driveInformationSAS_SATA) string {

	if driveInfo.rotationRate == 0x0000 {
		return "Not Available"
	}

	if driveInfo.rotationRate == 0x0001 {
		return "SSD"
	}
	return "HDD"
}
