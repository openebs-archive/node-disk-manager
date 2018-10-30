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

// SCSI command definitions.

package smart

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
)

// detectBusType detects the type of bus such as SCSI,Nvme,etc based on the device path given
func detectBusType(dname string) string {
	// define regex for identifying a nvme device
	RegexNVMeDev := regexp.MustCompile(`^/dev/nvme\d+n\d+$`)
	busType := "unknown"
	if strings.HasPrefix(dname, "/dev/sd") {
		busType = "SCSI"
	} else if strings.HasPrefix(dname, "/dev/hd") {
		busType = "IDE"
	} else if RegexNVMeDev.MatchString(dname) {
		busType = "NVMe"
	}

	return busType
}

// detectSCSIType returns the type of SCSI device such as ATA or SAS by sending
// an inquiry command based on the device path given to it.
func detectSCSIType(name string) (Dev, error) {
	device := SCSIDev{DevName: name}
	if err := device.Open(); err != nil {
		return nil, err
	}
	// send a scsi inquiry command to the given device
	SCSIInquiry, err := device.scsiInquiry()
	if err != nil {
		return nil, err
	}
	// Check if device is an ATA device (For an ATA device VendorIdentification value should be equal to ATA)
	// For ATA, return pointer to SATA device else return pointer to SCSI device interface
	if SCSIInquiry.VendorID == [8]byte{0x41, 0x54, 0x41, 0x20, 0x20, 0x20, 0x20, 0x20} {
		return &SATA{device}, nil
	}

	return &device, nil
}

// isReadDeviceCapacity16Supported checks if the SCSI READ CAPACITY(16)
// command is supported by the device.
func (d *SCSIDev) isReadDeviceCapacity16Supported() (bool, error) {
	respBuf := make([]byte, 20)

	cdb := CDB12{}
	cdb[0] = SCSIReportSupportedOpCodes
	cdb[1] = SCSIReportSupportedOpCodesSA
	cdb[2] = 0x02
	cdb[3] = SCSIReadCapacity16
	binary.BigEndian.PutUint16(cdb[4:], uint16(SCSIReadCapacitySA))
	binary.BigEndian.PutUint32(cdb[6:], uint32(len(respBuf)))

	if err := d.sendSCSICDB(cdb[:], &respBuf); err != nil {
		return false, err
	}

	switch respBuf[1] & 0x07 {
	case 0x03, 0x05:
		return true, nil
	default:
		return false, nil
	}
}

// readDeviceCapacity returns the devices capacity in bytes.
func (d *SCSIDev) readDeviceCapacity() (uint64, error) {
	if supported16, err := d.isReadDeviceCapacity16Supported(); err != nil {
		return 0, err
	} else if supported16 {
		return d.readDeviceCapacity16()
	} else {
		return d.readDeviceCapacity10()
	}
}

// readDeviceCapacity10 sends a SCSI READ CAPACITY(10) command to a device and
// returns the capacity in bytes.
func (d *SCSIDev) readDeviceCapacity10() (uint64, error) {
	respBuf := make([]byte, 8)

	// Use cdb10 to send a scsi read capacity command
	cdb := CDB10{SCSIReadCapacity10}

	// If sending scsi read capacity scsi command fails then return disk capacity
	// value 0 with error
	if err := d.sendSCSICDB(cdb[:], &respBuf); err != nil {
		return 0, err
	}

	lastLBA := binary.BigEndian.Uint32(respBuf[0:])          // max. addressable LBA
	LBsize := binary.BigEndian.Uint32(respBuf[4:])           // logical block size
	DeviceCapacity := (uint64(lastLBA) + 1) * uint64(LBsize) // calculate capacity

	return DeviceCapacity, nil
}

// readDeviceCapacity16 sends a SCSI READ CAPACITY(16) command to a device and
// returns the capacity in bytes.
func (d *SCSIDev) readDeviceCapacity16() (uint64, error) {
	respBuf := make([]byte, 32)

	// Use cdb16 to send a scsi read capacity command
	cdb := CDB16{SCSIReadCapacity16, SCSIReadCapacitySA}
	binary.BigEndian.PutUint32(cdb[10:], uint32(len(respBuf)))

	// If sending scsi read capacity scsi command fails then return disk capacity
	// value 0 with error
	if err := d.sendSCSICDB(cdb[:], &respBuf); err != nil {
		return 0, err
	}

	lastLBA := binary.BigEndian.Uint64(respBuf[0:])  // max. addressable LBA
	LBsize := binary.BigEndian.Uint32(respBuf[8:])   // logical block size
	DeviceCapacity := (lastLBA + 1) * uint64(LBsize) // calculate capacity

	return DeviceCapacity, nil
}

// isSatisfyCondition checks if the necessary conditions are met or not before getting
// the details for a particular disk such as binary permissions, bus type, etc
func isConditionSatisfied(devPath string) error {
	// Check if device path is given or not, if not provided then return with error
	if devPath == "" {
		return fmt.Errorf("no disk device path given to get the disk details")
	}

	// Check if required permissions are present or not for accessing a device
	if err := CheckBinaryPerm(); err != nil {
		return fmt.Errorf("error while checking device access permissions, Error: %+v", err)
	}

	// Check the type of bus such as SCSI, Nvme, etc
	busType := detectBusType(devPath)
	if busType != SupportedBusType {
		return fmt.Errorf("the device type is not supported yet, device type: %q", busType)
	}

	return nil
}
