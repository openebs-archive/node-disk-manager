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
	"unsafe"
)

// SCSI commands being used
const (
	SCSIModeSense    = 0x1a // mode sense command
	SCSIReadCapacity = 0x25 // read capacity command
	SCSIATAPassThru  = 0x85 // ata passthru command
)

// SCSI Command Descriptor Block types are the various type of scsi cdbs which are used
// to specify the various parameters or data required to send a particular scsi command

// CDB6 is an array of 6 byte
type CDB6 [6]byte

// CDB10 is an array of 10 byte
type CDB10 [10]byte

// CDB16 is an array of 16 byte
type CDB16 [16]byte

// getLBSize returns the logical block size of a SCSI device
func (d *SCSIDev) getLBSize() (uint32, error) {
	response := make([]byte, 8)
	// Use cdb16 to send a scsi read capacity command to get the
	// logical block size
	cdb := CDB16{SCSIReadCapacity}

	// If sending scsi read capacity scsi command fails then return
	// logical size value 0 with error
	if err := d.sendSCSICDB(cdb[:], &response); err != nil {
		return 0, err
	}
	LBsize := binary.BigEndian.Uint32(response[4:]) // logical block size

	return LBsize, nil
}

// runSCSIGen executes SCSI generic commands i.e sgIO commands
func (d *SCSIDev) runSCSIGen(header *sgIOHeader) error {
	// send an scsi generic ioctl command to the given file descriptor,
	// returns err if it fails to send it.

	if err := Ioctl(uintptr(d.fd), SGIO, uintptr(unsafe.Pointer(header))); err != nil {
		return err
	}
	// See http://www.t10.org/lists/2status.htm for SCSI status codes
	// TODO : Decode the status codes and return descriptive errors
	if header.info&SGInfoOkMask != SGInfoOk {
		err := sgIOErr{
			scsiStatus:   header.status,       // status code returned by an SCSI device
			hostStatus:   header.hostStatus,   // status code returned by a host
			driverStatus: header.driverStatus, // status code returned by a scsi driver
		}
		return err
	}

	return nil
}

// Error returns error string of error occured while sending SGIO(scsi generic ioctl)
// to a scsi device using the sgIOErr format
func (s sgIOErr) Error() string {
	return fmt.Sprintf("SCSI status: %#02x, host status: %#02x, driver status: %#02x",
		s.scsiStatus, s.hostStatus, s.driverStatus)
}

// sendSCSICDB sends a SCSI Command Descriptor Block to the device and writes the response into the
// supplied []byte pointer.
func (d *SCSIDev) sendSCSICDB(cdb []byte, respBuf *[]byte) error {
	senseBuf := make([]byte, 32)

	// Populate all the required fields of "sg_io_hdr_t" struct while sending
	// scsi command
	header := sgIOHeader{
		interfaceID:    'S',
		dxferDirection: SGDxferFromDev,
		cmdLen:         uint8(len(cdb)),
		mxSBLen:        uint8(len(senseBuf)),
		dxferLen:       uint32(len(*respBuf)),
		dxferp:         uintptr(unsafe.Pointer(&(*respBuf)[0])),
		cmdp:           uintptr(unsafe.Pointer(&cdb[0])),
		sbp:            uintptr(unsafe.Pointer(&senseBuf[0])),
		timeout:        DefaultTimeout,
	}

	return d.runSCSIGen(&header)
}

// modeSense function is used to send a SCSI MODE SENSE(6) command to a device.
// TODO : Implement SCSI MODE SENSE(10) command also
func (d *SCSIDev) modeSense(pageNo uint8, subPageNo uint8, pageCtrl uint8, disableBlockDesc bool) ([]byte, error) {
	respBuf := make([]byte, 64)
	var cdb1 uint8

	// if disable block descriptor value set to true then set cdb1 value
	// else it is 0
	if disableBlockDesc {
		cdb1 = (1 << 3)
	}

	// Populate all the required fields of cdb6 to send a scsi mode sense command
	cdb := CDB6{SCSIModeSense}
	cdb[0] = SCSIModeSense
	cdb[1] = cdb1
	cdb[2] = (pageCtrl << 6) | pageNo
	cdb[3] = subPageNo
	cdb[4] = uint8(len(respBuf))
	cdb[5] = 0

	// return error if sending mode sense command using cdb6 fails
	// TODO: Implement mode sense 10 and mode sense 16 also in order to
	// use them if sending mode sense 6 fails for a particular device for
	// a particular device page
	if err := d.sendSCSICDB(cdb[:], &respBuf); err != nil {
		return respBuf, err
	}

	return respBuf, nil
}
