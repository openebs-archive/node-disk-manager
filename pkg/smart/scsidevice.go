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

// SCSI generic IO functions.

package smart

import (
	"fmt"

	"github.com/openebs/node-disk-manager/pkg/util"
	"golang.org/x/sys/unix"
)

// Dev is the top-level device interface. All supported device
// types must implement these interfaces.
type Dev interface {
	DevOpen
	DevClose
	DevBasicinfoByAttr
	DevBasicDiskInfo
}

// DevOpen interface implements open method for opening a disk device
type DevOpen interface {
	Open() error
}

// DevClose interface implements close method for closing a disk device
type DevClose interface {
	Close() error
}

// DevBasicinfoByAttr interface implements getBasicDiskInfoByAttr method
// for getting particular attribute detail of a disk device
type DevBasicinfoByAttr interface {
	getBasicDiskInfoByAttr(attrName string) (string, error)
}

// DevBasicDiskInfo interface implements getBasicDiskInfo method for getting all the
// available details for a particular disk device
type DevBasicDiskInfo interface {
	getBasicDiskInfo() (DiskAttr, map[string]error)
}

// Open returns error if a SCSI device returns error when opened
func (d *SCSIDev) Open() (err error) {
	d.fd, err = unix.Open(d.DevName, unix.O_RDWR, 0600)
	return err
}

// Close returns error if a SCSI device is not closed
func (d *SCSIDev) Close() error {
	return unix.Close(d.fd)
}

// getBasicDiskInfoByAttr returns detail for a particular disk device attribute such as vendor,serial,etc
func (d *SCSIDev) getBasicDiskInfoByAttr(attrName string) (string, error) {
	// TODO : Return all the basic disk attributes available for a particular disk such
	// as rpm, etc

	// Check if given attribute can be fetched using SCSI Inquiry command
	if _, ok := ScsiInqAttr[attrName]; ok {
		// Get Attr value using SCSI inquiry command
		AttrValue, err := d.getAttrValUsingSI(attrName)
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

// getAttrValUsingSI returns the value of an attribute using standard SCSI Inquiry command
func (d *SCSIDev) getAttrValUsingSI(attrName string) (string, error) {
	// Standard SCSI INQUIRY command
	InqRes, err := d.scsiInquiry()
	if err != nil {
		return "", fmt.Errorf("error in sending SCSI Inquiry command, Error: %+v", err)
	}
	return InqRes.getValue()[attrName], nil
}

// getAttrValUsingSimpleSCSI returns the value of an attribute using implemented scsi
// commands other than SCSI inquiry command such as readCapacity.
func (d *SCSIDev) getAttrValUsingSimpleSCSI(attrName string) (string, error) {
	switch attrName {
	case LogicalSectorSize:
		LBSize, err := d.getLBSize()
		if err != nil {
			return "", fmt.Errorf("error in getting logical block size of the device, Error: %+v", err)
		}
		return fmt.Sprintf("%d", LBSize), nil
	case Capacity:
		capacity, err := d.readDeviceCapacity()
		if err != nil {
			return "", fmt.Errorf("error in getting total capacity of the device, Error: %+v", err)
		}
		return fmt.Sprintf("%d", capacity), nil
	default:
		return "", fmt.Errorf("Value of attribute %q not found", attrName)
	}
}

// getBasicDiskInfo returns all the available basic details for a particular disk device(except smart attr)
func (d *SCSIDev) getBasicDiskInfo() (DiskAttr, map[string]error) {
	// Collector will be used to collect errors in a string to error map
	collector := util.NewErrorCollector()
	collectedErrors := collector.Error()
	diskDetails := DiskAttr{}

	// Fill disk info using SCSI Inquiry command
	diskDetails, err := d.fillDiskInfoUsingSI(diskDetails)
	if err != nil {
		collector.Collect(SCSIInqErr, err)
	}

	// Scsi ReadCapacity command to get the capacity of a disk
	capacity, err := d.readDeviceCapacity()
	if !collector.Collect(SCSIReadCapErr, err) {
		diskDetails.Capacity = capacity
	}

	// SCSI ReadCapacity command to get the Logical block size of a scsi disk
	LBSize, err := d.getLBSize()
	if !collector.Collect(SCSiGetLBSizeErr, err) {
		diskDetails.LBSize = LBSize
	}

	return diskDetails, collectedErrors
}

// fillDiskInfoUsingSI fills the DiskAttr struct with the disk info fetched using SCSI
// Inquiry command
func (d *SCSIDev) fillDiskInfoUsingSI(diskDetails DiskAttr) (DiskAttr, error) {
	// Standard SCSI INQUIRY command
	InqRes, err := d.scsiInquiry()
	if err != nil {
		return diskDetails, err
	}
	diskDetails.SPCVersion = InqRes.getValue()[SPCVersion]
	diskDetails.Vendor = InqRes.getValue()[Vendor]
	diskDetails.ModelNumber = InqRes.getValue()[ModelNumber]
	diskDetails.FirmwareRevision = InqRes.getValue()[FirmwareRev]
	diskDetails.SerialNumber = InqRes.getValue()[SerialNumber]

	return diskDetails, nil
}
