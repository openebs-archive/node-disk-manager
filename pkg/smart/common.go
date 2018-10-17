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

// sendReadCap10 is used to send a SCSIReadCapacity10 scsi command to
// get logical block address and logical size of a disk
func (d *SCSIDev) sendReadCap10() (uint32, uint32, error) {
	respBuf10 := make([]byte, 8) // respBuf for read capacity(10) scsi cmd

	// Populate the CDB10 required fields to send SCSIReadCapacity10 command
	cdb10 := CDB16{SCSIReadCapacity10}

	// If sending scsi read capacity(10) command fails then return disk LBA
	// and LB Size value 0 with error
	if err := d.sendSCSICDB(cdb10[:], &respBuf10); err != nil {
		return 0, 0, err
	}

	lastLBA := binary.BigEndian.Uint32(respBuf10[0:]) // max. addressable LBA
	LBsize := binary.BigEndian.Uint32(respBuf10[4:])  // logical block size

	return lastLBA, LBsize, nil
}

// sendReadCap16 is used to send a SCSIReadCapacity16 scsi command to
// get logical block address and logical size of a disk
func (d *SCSIDev) sendReadCap16() (uint64, uint32, error) {
	respBuf16 := make([]byte, 32) // respBuf for read capacity(16) scsi cmd

	// Populate the CDB16 required fields to send SCSIReadCapacity16 command
	cdb16 := CDB16{SCSIReadCapacity16}
	cdb16[0] = SCSIReadCapacity16 // opcode for readCapacity(16) command
	cdb16[1] = SAReadCapacity16   // service action for readCapacity(16) command

	// If sending scsi read capacity(16) command fails then return disk LBA
	// and LB Size value 0 with error
	if err := d.sendSCSICDB(cdb16[:], &respBuf16); err != nil {
		return 0, 0, err
	}
	// longlastLBA will have the logical block size for the disk fetched
	// by sending SCSIReadCapacity16 command
	longlastLBA := binary.BigEndian.Uint64(respBuf16[0:]) // max. addressable LBA
	LBsize := binary.BigEndian.Uint32(respBuf16[8:])      // logical block size

	return longlastLBA, LBsize, nil
}

// readDeviceCapacity sends a SCSIReadCapacity10 or SCSIReadCapacity16 or both command
// to a device and returns the capacity in bytes.
func (d *SCSIDev) readDeviceCapacity() (uint64, error) {

	// DeviceCapacity will have the total capacity of a disk
	var DeviceCapacity uint64

	// First send a readcapacity10 command to get the disk logical block
	// address and logical size
	readCap10LBA, readCap10LBSize, err := d.sendReadCap10()
	if err != nil {
		return DeviceCapacity, err
	}

	// Check if the readCap10LBA size has reached the maximum value(0xffffffff) or not.
	// If it is equal to 0xffffffff(4294967295), then try SCSIReadCapacity16 command
	// to get the logical block address and logical size of the disk.
	if readCap10LBA != 4294967295 {
		// SCSIReadCapacity10 succeeded
		DeviceCapacity = (uint64(readCap10LBA) + 1) * uint64(readCap10LBSize)
	} else {
		// Since the logical block address(LBA) reported by SCSIReadCapacity10 has exceeded
		// the maximum limit, we will try SCSIReadCapacity16 to fetch the logical block address
		// and logical size of the disk.
		readCap16LBA, readCap16LBSize, err := d.sendReadCap16()
		if err != nil {
			return DeviceCapacity, err
		}
		DeviceCapacity = (readCap16LBA + 1) * uint64(readCap16LBSize)
	}

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
