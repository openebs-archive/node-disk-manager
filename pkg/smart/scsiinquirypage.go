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
	"bytes"
	"encoding/binary"
	"fmt"
)

// commands used to fetch various informations of a disk from a set of defined
// scsi pages
const (
	SCSIInquiry = 0x12 // inquiry command

	// Minimum length of standard INQUIRY response
	INQRespLen = 56
)

// scsiInquiry sends an INQUIRY command to a SCSI device and returns an
// InquiryResponse struct.
func (d *SCSIDev) scsiInquiry() (InquiryResponse, error) {
	var response InquiryResponse
	respBuf := make([]byte, INQRespLen)

	// Use cdb6 to send scsi inquiry command
	// TODO: Use cdb10 or cdb16 also as a fallthrough command to if
	// cdb6 fails for a particular device
	cdb := CDB6{SCSIInquiry}

	binary.BigEndian.PutUint16(cdb[3:], uint16(len(respBuf)))

	// return error if sending scsi cdb6 command fails
	if err := d.sendSCSICDB(cdb[:], &respBuf); err != nil {
		return response, err
	}

	binary.Read(bytes.NewBuffer(respBuf), NativeEndian, &response)

	return response, nil
}

// getValue returns the value of the attributes fetched using inquiry
// command using Inquiry response struct
func (inquiry InquiryResponse) getValue() map[string]string {
	InqRespMap := make(map[string]string)

	SPCVersionValue := fmt.Sprintf("%.d", (inquiry.Version - 0x02))

	InqRespMap[Compliance] = "SPC-" + SPCVersionValue
	InqRespMap[Vendor] = fmt.Sprintf("%.8s", inquiry.VendorID)
	InqRespMap[ModelNumber] = fmt.Sprintf("%.16s", inquiry.ProductID)
	InqRespMap[FirmwareRev] = fmt.Sprintf("%.4s", inquiry.ProductRev)
	InqRespMap[SerialNumber] = fmt.Sprintf("%.20s", inquiry.SerialNumber)
	return InqRespMap
}
