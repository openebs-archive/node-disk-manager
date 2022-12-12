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

/*
Package smart provides details about a particular disk device which includes both
basic details such as vendor, model, serial, etc as well as smart details such as
Raw_Read_Error_Rate, Temperature_Celsius, Spin_Up_Time, etc by parsing various disk
pages such as inquiry page, ata command set page ,etc using various SCSI commands such
as scsi inquiry, read device capacity, mode sense, etc.

NOTE : For now, the implementation is only for getting the basic details (not smart details)
of SCSI disks such as vendor,serial, model, firmware revision, logical sector size,etc.

Usage:

	import "github.com/openebs/node-disk-manager/pkg/smart"

S.M.A.R.T. (Self-Monitoring, Analysis and Reporting Technology; often written as SMART) is
a monitoring system included in computer hard disk drives (HDDs), solid-state drives (SSDs),
and eMMC drives. Its primary function is to detect and report various indicators of drive
reliability with the intent of anticipating imminent hardware failures.
When S.M.A.R.T. data indicates a possible imminent drive failure, software running on the
host system may notify the user so preventative action can be taken to prevent data loss,
and the failing drive can be replaced and data integrity maintained.

This smart go library provides functionality to get the list of all the SCSI disk devices
attached to it and a list of disk smart attributes along with the basic disk attributes
such as serial no, sector size, wwn, rpm, vendor, etc.

For getting the basic SCSI disk details, one should import this library and then a call to
function SCSIBasicDiskInfo("device-path") where device-path is the devpath of the scsi device
e.g. /dev/sda, /dev/sdb, etc for which we want to fetch the basic disk details such as vendor,
model, serial, wwn, logical and physical sector size, etc is to be fetched.
This function would return a struct of disk details alongwith errors if any, filled by smart
library for the particular disk device whose devpath has been given.

If a user wants to only get the detail of a particular disk attribute such as vendor, serial,
etc for a particular SCSI device then function SCSIBasicDiskInfoByAttrName(attrName string)
(where attrName refers to the attribute whose value is to be fetched such as Vendor)
should be called which will return the detail or value of that particular attribute only
alongwith the errors if any occurred while fetching the detail.

An example usage can be like this -

package smartusageexample

import (

	"fmt"

	"k8s.io/klog/v2"
	"github.com/openebs/node-disk-manager/pkg/smart"

)

func main() {
	deviceBasicSCSIInfo, err := smart.SCSIBasicDiskInfo("/dev/sda")
	if err != nil {
		klog.Fatal(err)
	}

	fmt.Printf("Vendor :%s \n",deviceBasicSCSIInfo.Vendor)
	fmt.Printf("Compliance :%s \n",deviceBasicSCSIInfo.Compliance)
	fmt.Printf("FirmwareRevision :%s \n",deviceBasicSCSIInfo.FirmwareRevision)
	fmt.Printf("Capacity :%d \n",deviceBasicSCSIInfo.Capacity)
}

NOTE : This document will remain in continuous updation whenever more features and
functionalities are implemented.

Please refer to the design doc here -
https://docs.google.com/document/d/1avZrFI3j1AOmWIY_43oyK9Nkj5IYT37fzIhAAbp0Bxs/edit?usp=sharing
*/
package smart
