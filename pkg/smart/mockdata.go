/*
Copyright 2018 OpenEBS Authors.

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

package smart

import "github.com/openebs/node-disk-manager/pkg/udev"

type MockOsDiskDetails struct {
	SPCVersion       string
	FirmwareRevision string
	Capacity         uint64
	LBSize           uint32
	DevPath          string
}

type mockDev interface {
	Open() error
	Close() error
	mockDiskDetailsBySmart() (MockOsDiskDetails, error)
	getCommonSCSIDetails(cDetail MockOsDiskDetails) (MockOsDiskDetails, error)
}

var (
	diskDetails MockOsDiskDetails
)

func (d *SCSIDev) getCommonSCSIDetails(cDetail MockOsDiskDetails) (MockOsDiskDetails, error) {
	InqRes, err := d.scsiInquiry()
	if err != nil {
		return cDetail, err
	}
	cDetail.SPCVersion = InqRes.getValue()[SPCVersion]
	cDetail.FirmwareRevision = InqRes.getValue()[FirmwareRev]

	// Scsi readDeviceCapacity command to get the capacity of a disk
	capacity, err := d.readDeviceCapacity()
	if err != nil {
		return cDetail, err
	}
	cDetail.Capacity = capacity

	devPath, _ := getDevPath()
	cDetail.DevPath = devPath

	return cDetail, nil
}

func (d *SATA) mockDiskDetailsBySmart() (MockOsDiskDetails, error) {

	diskDetails, err := d.getCommonSCSIDetails(diskDetails)
	if err != nil {
		return diskDetails, err
	}
	identifyBuf, err := d.ataIdentify()
	if err != nil {
		return diskDetails, err
	}
	LBSize, _ := identifyBuf.getSectorSize()
	diskDetails.LBSize = LBSize

	return diskDetails, nil

}

func (d *SCSIDev) mockDiskDetailsBySmart() (MockOsDiskDetails, error) {

	diskDetails, err := d.getCommonSCSIDetails(diskDetails)
	if err != nil {
		return diskDetails, err
	}

	LBSize, err := d.getLBSize()
	if err != nil {
		return diskDetails, err
	}
	diskDetails.LBSize = LBSize

	return diskDetails, nil

}

// MockScsiBasicDiskInfo is used to fetch basic disk details for a scsi disk
func MockScsiBasicDiskInfo() (MockOsDiskDetails, error) {

	osname, err := udev.OsDiskName()
	if err != nil {
		return diskDetails, err
	}
	devPath := "/dev/" + osname
	// Before getting disk details, check if the necessary conditions to get
	// disk details are fulfilled or not such as device path is given or not,
	// binary permissions are set or not, bus type is supported or not, etc
	if err := isConditionSatisfied(devPath); err != nil {
		return diskDetails, err
	}
	// Check the type of SCSI device, if it is ATA or something else..
	// based on which the dev interface is returned
	d, err := mockdetectSCSIType(devPath)
	if err != nil {
		return diskDetails, err
	}
	defer d.Close()

	// Get all the available disk details in the form of struct and errors if any (in form of map)
	diskDetails, err = d.mockDiskDetailsBySmart()
	if err != nil {
		return diskDetails, err
	}

	return diskDetails, nil
}

func mockdetectSCSIType(name string) (mockDev, error) {
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

func getDevPath() (string, error) {
	osname, err := udev.OsDiskName()
	if err != nil {
		return "", err
	}
	devPath := "/dev/" + osname

	return devPath, nil
}
