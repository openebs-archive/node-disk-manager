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

// Functions for SCSI-ATA Translation.

package smart

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/openebs/node-disk-manager/pkg/util"
)

// SATA is a simple wrapper around an embedded SCSIDevice type, which handles sending ATA
// commands via SCSI pass-through (SCSI-ATA Translation).
type SATA struct {
	SCSIDev
}

// ataIdentify sends SCSI_ATA_PASSTHRU_16 command and read data from the response
// received based on the defined ATA IDENTIFY STRUCT
func (d *SATA) ataIdentify() (ATACSPage, error) {
	var identifyBuf ATACSPage
	responseBuf := make([]byte, 512)

	// SCSI_ATA_PASSTHRU_16 command
	// Here, we are only populating the required fields to send a scsi cdb16 command
	// to an ata device such as cdb16[1], cdb16[2], etc since the rest are optional
	// and needs to be filled only for some specific purposes or constraints to apply
	// while getting the information
	cdb16 := CDB16{SCSIATAPassThru}
	cdb16[1] = 0x08               // ATA protocol (4 << 1, PIO data-in)
	cdb16[2] = 0x0e               // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb16[14] = AtaIdentifyDevice // miscellaneous CDB information

	// Send a CDB16 type scsi command to an ata device
	if err := d.sendSCSICDB(cdb16[:], &responseBuf); err != nil {
		return identifyBuf, fmt.Errorf("error in sending SCSICDB 16 for ATA device, Error: %+v", err)
	}
	binary.Read(bytes.NewBuffer(responseBuf), util.NativeEndian, &identifyBuf)
	return identifyBuf, nil
}

// getBasicDiskInfoByAttr returns all the disk attributes and smart info for a particular SATA device
func (d *SATA) getBasicDiskInfoByAttr(attrName string) (string, error) {
	// Check if given attribute can be fetched using SCSI Inquiry command
	if _, ok := ScsiInqAttr[attrName]; ok {
		// Get Attr value using SCSI inquiry command
		AttrValue, err := d.getAttrValUsingSI(attrName)
		if err != nil {
			return "", err
		}
		return AttrValue, nil
	}

	// Check if given attribute can be fetched using ATA8-ACS page
	if _, ok := ATACSAttr[attrName]; ok {
		// Get Attr value using ATA8-ACS ATA command set page
		AttrValue, err := d.getAttrValUsingATACSPage(attrName)
		if err != nil {
			return "", err
		}
		return AttrValue, nil
	}

	// Check if given attribute can be fetched using simple scsi commands such as readcapacity
	if _, ok := SimpleSCSIAttr[attrName]; ok {
		// Get Attr value using implemented SCSI commands other than SCSI Inquiry command
		AttrValue, err := d.getAttrValUsingSimpleSCSI(attrName)
		if err != nil {
			return "", err
		}
		return AttrValue, nil
	}
	return "", fmt.Errorf("Value of attribute %q not found", attrName)
}

func (d *SATA) getAttrValUsingATACSPage(attrName string) (string, error) {
	// store data from the response received by calling ataIdentify based on the defined ATA IDENTIFY STRUCT
	identifyBuf, err := d.ataIdentify()
	if err != nil {
		return "", fmt.Errorf("error in sending ATAIdentifyCommand, Error: %+v", err)
	}
	switch attrName {
	case WWN:
		return identifyBuf.getWWN(), nil
	case AtATransport:
		return identifyBuf.ataTransport(), nil
	case ATAMajor:
		return identifyBuf.getATAMajorVersion(), nil
	case ATAMinor:
		return identifyBuf.getATAMinorVersion(), nil
	case RPM:
		return fmt.Sprintf("%d", identifyBuf.RotationRate), nil
	case LogicalSectorSize:
		LBSize, _ := identifyBuf.getSectorSize()
		return fmt.Sprintf("%d", LBSize), nil
	case PhysicalSectorSize:
		_, PBSize := identifyBuf.getSectorSize()
		return fmt.Sprintf("%d", PBSize), nil
	default:
		return "", fmt.Errorf("Value of attribute %q not found", attrName)
	}
}

// getBasicDiskInfo returns all the available basic details for a particular disk device(except smart attr)
func (d *SATA) getBasicDiskInfo() (DiskAttr, map[string]error) {
	// Collector will be used to collect errors in a string to error map
	collector := util.NewErrorCollector()
	collectedErrors := collector.Error()
	diskDetails := DiskAttr{}

	// Fill disk info using ATA command set page defined as ATA8-ACS
	diskDetails, err := d.fillDiskInfoUsingATACSPage(diskDetails)
	if err != nil {
		collector.Collect(ATAIdentifyErr, err)
	}
	// Fill disk info using SCSI Inquiry
	diskDetails, err = d.fillDiskInfoUsingSI(diskDetails)
	if err != nil {
		collector.Collect(SCSIInqErr, err)
	}
	// Scsi readDeviceCapacity command to get the capacity of a disk
	capacity, err := d.readDeviceCapacity()
	if !collector.Collect(SCSIReadCapErr, err) {
		diskDetails.Capacity = capacity
	}

	return diskDetails, collectedErrors
}

// fillDiskInfoUsingATACSPage fills the disk information by parsing ATA8-ACS command set
// page for a particular ATA device
func (d *SATA) fillDiskInfoUsingATACSPage(diskDetails DiskAttr) (DiskAttr, error) {
	// store data from the response received by calling ataIdentify based on the defined ATA IDENTIFY STRUCT
	identifyBuf, err := d.ataIdentify()
	if err != nil {
		return diskDetails, err
	}
	LBSize, PBSize := identifyBuf.getSectorSize()

	diskDetails.SerialNumber = string(identifyBuf.getSerialNumber())
	diskDetails.WWN = identifyBuf.getWWN()
	diskDetails.LBSize = LBSize
	diskDetails.PBSize = PBSize
	diskDetails.RotationRate = identifyBuf.RotationRate
	diskDetails.AtaTransport = identifyBuf.ataTransport()
	diskDetails.ATAMajorVersion = identifyBuf.getATAMajorVersion()
	diskDetails.ATAMinorVersion = identifyBuf.getATAMinorVersion()

	return diskDetails, nil
}
