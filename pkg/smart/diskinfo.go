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

// SCSIBasicDiskInfo returns all the available disk details for a particular disk device
func (I *Identifier) SCSIBasicDiskInfo() (DiskAttr, map[string]error) {
	diskDetail := DiskAttr{}
	// Collector will be used to collect errors in a string to error map
	collector := NewErrorCollector()

	// Before getting disk details, check if the necessary conditions to get
	// disk details are fulfilled or not such as device path is given or not,
	// binary permissions are set or not, bus type is supported or not, etc
	if err := isConditionSatisfied(I.DevPath); err != nil {
		collector.Collect(errorCheckConditions, err)
		return diskDetail, collector.Error()
	}
	// Check the type of SCSI device, if it is ATA or something else..
	// based on which the dev interface is returned
	d, err := detectSCSIType(I.DevPath)
	if collector.Collect(DetectSCSITypeErr, err) {
		return diskDetail, collector.Error()
	}
	defer d.Close()

	// Get all the available disk details in the form of struct and errors if any (in form of map)
	diskDetail, errorMap := d.getBasicDiskInfo()

	return diskDetail, errorMap
}

// SCSIBasicDiskInfoByAttrName returns disk details(disk attributes and their values such as vendor,serialno,etc) of a disk
func (I *Identifier) SCSIBasicDiskInfoByAttrName(attrName string) (string, error) {
	var AttrDetail string
	// Before getting disk details, check if the necessary conditions to get
	// disk details are fulfilled or not
	if err := isConditionSatisfied(I.DevPath); err != nil {
		return "", err
	}

	// Check if any attribute is given or not
	if attrName == "" {
		return "", fmt.Errorf("no attribute name specified to get the value")
	}

	// Check the type of SCSI device, if it is ATA or something else..
	// based on which the dev interface is returned
	d, err := detectSCSIType(I.DevPath)
	if err != nil {
		return "", fmt.Errorf("error in detecting type of SCSI device, Error: %+v", err)
	}
	defer d.Close()

	// Get the value of a particular disk attribute
	AttrDetail, err = d.getBasicDiskInfoByAttr(attrName)
	if err != nil {
		return AttrDetail, fmt.Errorf("error getting %q of disk having devpath %q, error: %+v", attrName, I.DevPath, err)
	}
	return AttrDetail, nil
}
