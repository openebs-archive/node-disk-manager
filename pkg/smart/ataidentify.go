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

import (
	"fmt"
)

// swapByteOrder swaps the order of every second byte in a byte slice and
// modifies the slice in place.
func (d *ATACSPage) swapByteOrder(b []byte) []byte {
	tmp := make([]byte, len(b))

	for i := 0; i < len(b); i += 2 {
		tmp[i], tmp[i+1] = b[i+1], b[i]
	}
	return tmp
}

// getSerialNumber returns the serial number of a disk device
// using an ATA IDENTIFY command.
func (d *ATACSPage) getSerialNumber() []byte {
	// Needed a byte swap since each pair of bytes in an ATA string is swapped
	// word 0 | offset 0 -> second character, offset 1 -> First character and so on.
	return d.swapByteOrder(d.SerialNumber[:])
}

// getWWN returns the worldwide unique name for a disk
// The World Wide Name (WWN) uses the NAA IEEE Registered designator format
// defined in SPC-5 with the NAA field set to 5h.
func (d *ATACSPage) getWWN() string {
	// NAA field indicates the format of the world wide name
	NAA := d.WWN[0] >> 12
	// IEEE OUI field shall contain the 24-bit canonical form company
	// identifier (i.e., OUI) that the IEEE has assigned to the device manufacturer
	IEEEOUI := (uint32(d.WWN[0]&0x0fff) << 12) | (uint32(d.WWN[1]) >> 4)
	// UNIQUE ID field shall contain a value assigned by the device manufacturer
	// that is unique for the device within the OUI domain
	UniqueID := ((uint64(d.WWN[1]) & 0xf) << 32) | (uint64(d.WWN[2]) << 16) | uint64(d.WWN[3])

	return fmt.Sprintf("%x %06x %09x", NAA, IEEEOUI, UniqueID)
}

// getSectorSize returns logical and physical sector sizes of a disk
func (d *ATACSPage) getSectorSize() (uint32, uint32) {
	// By default, we are assuming the physical and logical sector size to be 512
	// based on further check conditions, it would be altered.
	var LogSec, PhySec uint32 = 512, 512

	if (d.SectorSize & 0xc000) != 0x4000 {
		return LogSec, PhySec
	}
	// TODO : Add support for Long Logical/Physical Sectors (LLS/LPS)
	if (d.SectorSize & 0x2000) != 0x0000 {
		// Physical sector size is multiple of logical sector size
		PhySec <<= (d.SectorSize & 0x0f)
	}
	return LogSec, PhySec
}

// getATAMajorVersion returns the ATA major version of a disk using an ATA IDENTIFY command.
func (d *ATACSPage) getATAMajorVersion() (s string) {
	if (d.MajorVer == 0) || (d.MajorVer == 0xffff) {
		return "This device does not report ATA major version"
	}
	// ATA Major version word is a bitmask, hence we will get the most significant bit
	// of this word and then do a map lookup
	majorVer := MSignificantBit(uint(d.MajorVer))
	if s, ok := ataMajorVersions[majorVer]; ok {
		return s
	}
	return "unknown"
}

// getATAMinorVersion returns the ATA minor version using an ATA IDENTIFY command.
func (d *ATACSPage) getATAMinorVersion() string {
	if (d.MinorVer == 0) || (d.MinorVer == 0xffff) {
		return "This device does not report ATA minor version"
	}
	// Since ATA minor version word is not a bitmask, we simply do a map lookup here
	if s, ok := ataMinorVersions[d.MinorVer]; ok {
		return s
	}
	return "unknown"
}

// ataTransport returns the type of ata Transport being
// used such as serial ATA, parallel ATA, etc.
func (d *ATACSPage) ataTransport() (s string) {
	if (d.AtaTransportMajor == 0) || (d.AtaTransportMajor == 0xffff) {
		s = "This device does not report Transport"
		return
	}
	switch d.AtaTransportMajor >> 12 {
	case 0x0:
		s = "Parallel ATA" // parallel ata transport
	case 0x1:
		s = d.IdentifySerialATAType() // identify the type of serial ata as it is a serial ata transport
	case 0xe:
		s = fmt.Sprintf("PCIe (%#03x)", d.AtaTransportMajor&0x0fff)
	default:
		s = fmt.Sprintf("Unknown (%#04x)", d.AtaTransportMajor)
	}
	return
}

// IdentifySerialATAType identifies the type of SATA transport being used by a disk
func (d *ATACSPage) IdentifySerialATAType() (s string) {
	s = "Serial ATA"
	// Get the most significant bit of the ata transport major word as it is a bitmask
	// and then get the serial ata transport type based on the value
	transportMajor := MSignificantBit(uint(d.AtaTransportMajor & 0x0fff))
	// Lookup in the map for the type of serial ata transport based on key
	if serialATAType, ok := serialATAType[transportMajor]; ok {
		s += serialATAType
		return s
	}
	s += fmt.Sprintf(" SATA (%#03x)", d.AtaTransportMajor&0x0fff)
	return s
}
