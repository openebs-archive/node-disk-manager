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

package smart

// SCSI generic (sg)
// See dxfer_direction http://sg.danny.cz/sg/p/sg_v3_ho.html
const (
	SGDxferNone      = -1     //SCSI Test Unit Ready command
	SGDxferToDev     = -2     //SCSI WRITE command
	SGDxferFromDev   = -3     //SCSI READ command
	SGDxferToFromDev = -4     //relevant to indirect IO (otherwise it is treated like SGDxferFromDev)
	SGInfoOk         = 0x0    //no sense, host nor driver "noise" or error
	SGInfoOkMask     = 0x1    //indicates whether some error or status field is non-zero
	SGIO             = 0x2285 //scsi generic ioctl command
	DefaultTimeout   = 20000  //DefaultTimeout in millisecs
)

// ATA command being used
const (
	AtaIdentifyDevice = 0xec
)

// Constants being used by switch case for returning disk details
const (
	Compliance         = "Compliance"
	Vendor             = "Vendor"
	Capacity           = "Capacity"
	LogicalSectorSize  = "LogicalSectorSize"
	PhysicalSectorSize = "PhysicalSectorSize"
	SerialNumber       = "SerialNumber"
	WWN                = "LuWWNDeviceID"
	FirmwareRev        = "FirmwareRevision"
	ModelNumber        = "ModelNumber"
	RPM                = "RPM"
	ATAMajor           = "ATAMajorVersion"
	ATAMinor           = "ATAMinorVersion"
	AtATransport       = "AtaTransport"
	SupportedBusType   = "SCSI"
)

// Constants being used as keys for sending map of errors
const (
	SCSIInqErr           = "SCSIInquiryError"
	SCSIReadCapErr       = "SCSIReadcapacityError"
	ATAIdentifyErr       = "AtaIdentifyError"
	RPMErr               = "RPMError"
	SCSiGetLBSizeErr     = "GetLogicalBlockSizeError"
	DetectSCSITypeErr    = "DetectScsiTypeError"
	errorCheckConditions = "errorCheckingConditions"
)

// ataMajorVersions are the major versions defined for an ATA device in ATA command set page
// Table 9 of X3T13/2008D (ATA-3) Revision 7b, See http://www.scs.stanford.edu/11wi-cs140/pintos/specs/ata-3-std.pdf
// Table 29 of T13/1699-D Revision 6a, See http://www.t13.org/documents/uploadeddocuments/docs2008/d1699r6a-ata8-acs.pdf
// Table 45 of T13/2161-D Revision 5, See http://www.t13.org/Documents/UploadedDocuments/docs2013/d2161r5-ATAATAPI_Command_Set_-_3.pdf
// Table 55 of T13/BSR INCITS 529 Revision 18 , See http://t13.org/Documents/UploadedDocuments/docs2017/di529r18-ATAATAPI_Command_Set_-_4.pdf
var ataMajorVersions = map[int]string{
	1:  "ATA-1",       //obsolete
	2:  "ATA-2",       //obsolete
	3:  "ATA-3",       //obsolete
	4:  "ATA-ATAPI-4", //obsolete
	5:  "ATA-ATAPI-5",
	6:  "ATA-ATAPI-6",
	7:  "ATA-ATAPI-7",
	8:  "ATA8-ACS",
	9:  "ACS-2",
	10: "ACS-3",
	11: "ACS-4",
}

// Table 10 of X3T13/2008D (ATA-3) Revision 7b, See http://www.scs.stanford.edu/11wi-cs140/pintos/specs/ata-3-std.pdf
// Table 31 of T13/1699-D Revision 6a, See http://www.t13.org/documents/uploadeddocuments/docs2008/d1699r6a-ata8-acs.pdf
// Table 47 of T13/2161-D Revision 5, See http://www.t13.org/Documents/UploadedDocuments/docs2013/d2161r5-ATAATAPI_Command_Set_-_3.pdf
// Table 57 of T13/BSR INCITS 529 Revision 18 , See http://t13.org/Documents/UploadedDocuments/docs2017/di529r18-ATAATAPI_Command_Set_-_4.pdf
var ataMinorVersions = map[uint16]string{
	0x0001: "ATA-1 X3T9.2/781D prior to revision 4",       // obsolete
	0x0002: "ATA-1 published, ANSI X3.221-1994",           // obsolete
	0x0003: "ATA-1 X3T9.2/781D revision 4",                // obsolete
	0x0004: "ATA-2 published, ANSI X3.279-1996",           // obsolete
	0x0005: "ATA-2 X3T10/948D prior to revision 2k",       // obsolete
	0x0006: "ATA-3 X3T10/2008D revision 1",                // obsolete
	0x0007: "ATA-2 X3T10/948D revision 2k",                // obsolete
	0x0008: "ATA-3 X3T10/2008D revision 0",                // obsolete
	0x0009: "ATA-2 X3T10/948D revision 3",                 // obsolete
	0x000a: "ATA-3 published, ANSI X3.298-1997",           // obsolete
	0x000b: "ATA-3 X3T10/2008D revision 6",                // obsolete
	0x000c: "ATA-3 X3T13/2008D revision 7 and 7a",         // obsolete
	0x000d: "ATA/ATAPI-4 X3T13/1153D version 6",           // obsolete
	0x000e: "ATA/ATAPI-4 T13/1153D version 13",            // obsolete
	0x000f: "ATA/ATAPI-4 X3T13/1153D version 7",           // obsolete
	0x0010: "ATA/ATAPI-4 T13/1153D version 18",            // obsolete
	0x0011: "ATA/ATAPI-4 T13/1153D version 15",            // obsolete
	0x0012: "ATA/ATAPI-4 published, ANSI NCITS 317-1998",  // obsolete
	0x0013: "ATA/ATAPI-5 T13/1321D version 3",             // obsolete
	0x0014: "ATA/ATAPI-4 T13/1153D version 14",            // obsolete
	0x0015: "ATA/ATAPI-5 T13/1321D revision 1",            // obsolete
	0x0016: "ATA/ATAPI-5 published, ANSI NCITS 340-2000",  // obsolete
	0x0017: "ATA/ATAPI-4 T13/1153D revision 17",           // obsolete
	0x0018: "ATA/ATAPI-6 T13/1410D version 0",             // obsolete
	0x0019: "ATA/ATAPI-6 T13/1410D version 3a",            // obsolete
	0x001a: "ATA/ATAPI-7 T13/1532D version 1",             // obsolete
	0x001b: "ATA/ATAPI-6 T13/1410D version 2",             // obsolete
	0x001c: "ATA/ATAPI-6 T13/1410D version 1",             // obsolete
	0x001d: "ATA/ATAPI-7 published, ANSI INCITS 397-2005", // obsolete
	0x001e: "ATA/ATAPI-7 T13/1532D version 0",             // obsolete
	0x001f: "ACS-3 revision 3b",
	0x0021: "ATA/ATAPI-7 T13/1532D version 4a",            // obsolete
	0x0022: "ATA/ATAPI-6 published, ANSI INCITS 361-2002", // obsolete
	0x0027: "ATA8-ACS version 3c",
	0x0028: "ATA8-ACS version 6",
	0x0029: "ATA8-ACS version 4",
	0x0031: "ACS-2 revision 2",
	0x0033: "ATA8-ACS version 3e",
	0x0039: "ATA8-ACS version 4c",
	0x0042: "ATA8-ACS version 3f",
	0x0052: "ATA8-ACS version 3b",
	0x005e: "ACS-4 revision 5",
	0x006d: "ACS-3 revision 5",
	0x0082: "ACS-2 published, ANSI INCITS 482-2012",
	0x0107: "ATA8-ACS version 2d",
	0x010a: "ACS-3 published, ANSI INCITS 522-2014",
	0x0110: "ACS-2 revision 3",
	0x011b: "ACS-3 revision 4",
}

// serialATAType contains the various types of serial ata transport
var serialATAType = map[int]string{
	0: " ATA8-AST",
	1: " SATA 1.0a",
	2: " SATA II Ext",
	3: " SATA 2.5",
	4: " SATA 2.6",
	5: " SATA 3.0",
	6: " SATA 3.1",
	7: " SATA 3.2",
}

// ScsiInqAttr is the list of attributes fetched by SCSI Inquiry command
var ScsiInqAttr = map[string]bool{
	Compliance:   true,
	Vendor:       true,
	SerialNumber: true,
	ModelNumber:  true,
	FirmwareRev:  true,
}

// SimpleSCSIAttr is the list of attributes fetched by simple SCSI
// commands such as readCapacity,etc
var SimpleSCSIAttr = map[string]bool{
	PhysicalSectorSize: true,
	LogicalSectorSize:  true,
	Capacity:           true,
}

// ATACSAttr is the list of attributes fetched using ATACSPage
var ATACSAttr = map[string]bool{
	WWN:                true,
	AtATransport:       true,
	ATAMajor:           true,
	ATAMinor:           true,
	RPM:                true,
	LogicalSectorSize:  true,
	PhysicalSectorSize: true,
}

// Identifier (devPath such as /dev/sda,etc) is an identifier for smart probe
type Identifier struct {
	DevPath string
}

// ATACSPage struct is an ATA IDENTIFY DEVICE struct. ATA8-ACS defines this as a page of 16-bit words.
// _ (underscore) is used here to skip the words which we don't want to parse or get the data while parsing
// the ata identify device data struct page.
type ATACSPage struct {
	_                 [10]uint16  // ...
	SerialNumber      [20]byte    // Word 10..19, device serial number.
	_                 [60]uint16  // ...
	MajorVer          uint16      // Word 80, major version number.
	MinorVer          uint16      // Word 81, minor version number.
	_                 [24]uint16  // ...
	SectorSize        uint16      // Word 106, Logical/physical sector size.
	_                 [1]uint16   // ...
	WWN               [4]uint16   // Word 108..111, WWN (World Wide Name).
	_                 [105]uint16 // ...
	RotationRate      uint16      // Word 217, nominal media rotation rate.
	_                 [4]uint16   // ...
	AtaTransportMajor uint16      // Word 222, Transport major version number.
	_                 [33]uint16  // ...
} // 512 bytes

// SCSIDev represents a particular scsi device with device name
// and file descriptor
type SCSIDev struct {
	DevName string // SCSI device name
	fd      int    // File descriptor for the scsi device
}

// sg_io_hdr_t structure See http://sg.danny.cz/sg/p/sg_v3_ho.html
type sgIOHeader struct {
	interfaceID    int32   // 'S' for SCSI generic (required)
	dxferDirection int32   // data transfer direction
	cmdLen         uint8   // SCSI command length (<= 16 bytes)
	mxSBLen        uint8   // max length to write to sbp
	iovecCount     uint16  // 0 implies no scatter gather
	dxferLen       uint32  // byte count of data transfer
	dxferp         uintptr // points to data transfer memory or scatter gather list
	cmdp           uintptr // points to command to perform
	sbp            uintptr // points to sense_buffer memory
	timeout        uint32  // MAX_UINT -> no timeout (unit: millisec)
	flags          uint32  // 0 -> default, see SG_FLAG...
	packID         int32   // unused internally (normally)
	usrPtr         uintptr // unused internally
	status         uint8   // SCSI status
	maskedStatus   uint8   // shifted, masked scsi status
	msgStatus      uint8   // messaging level data (optional)
	SBLenwr        uint8   // byte count actually written to sbp
	hostStatus     uint16  // errors from host adapter
	driverStatus   uint16  // errors from software driver
	resid          int32   // dxfer_len - actual_transferred
	duration       uint32  // time taken by cmd (unit: millisec)
	info           uint32  // auxiliary information
}

// sgIOErr is the format in which error could be returned while sending
// a scsi generic ioctl command
type sgIOErr struct {
	scsiStatus   uint8
	hostStatus   uint16
	driverStatus uint16
}

// DiskAttr is struct being used for returning all the available disk details (both basic and smart)
// For now, only basic disk attr are being fetched so it is returning only basic attrs
type DiskAttr struct {
	BasicDiskAttr
	ATADiskAttr
}

// BasicDiskAttr is the structure being used for returning basic disk details
type BasicDiskAttr struct {
	Compliance       string
	Vendor           string
	ModelNumber      string
	SerialNumber     string
	FirmwareRevision string
	WWN              string
	Capacity         uint64
	LBSize           uint32
	PBSize           uint32
	RotationRate     uint16
}

// SmartDiskAttr is the structure defined for smart disk attrs (Note : Not being used yet)
type SmartDiskAttr struct {
}

// ATADiskAttr is the struct for disk attributes that are specific to ATA disks
type ATADiskAttr struct {
	ATAMajorVersion string
	ATAMinorVersion string
	AtaTransport    string
}

// InquiryResponse is used for parsing response fetched
// by sending a scsi inquiry command to a scsi device
// Here underscore (_) is used to skip the words which we don't want to
// parse as of now..
type InquiryResponse struct {
	_            [2]byte  // ...
	Version      byte     // implemented specification version such as SPC-1,SPC-2,etc
	_            [5]byte  // ...
	VendorID     [8]byte  // Vendor Identification
	ProductID    [16]byte // Product Identification
	ProductRev   [4]byte  // Product Revision Level
	SerialNumber [20]byte // Serial Number
}
