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

package v1

//BlockDeviceInfo exposes the json output of lsblk:"blockdevices"
type BlockDeviceInfo struct {
	Blockdevices []Blockdevice `json:"blockdevices"`
}

//Blockdevices has block disk fields
type Blockdevice struct {
	Name       string        `json:"name"`               //block device name
	Kname      string        `json:"kname"`              //internal kernel device nam
	Majmin     string        `json:"maj:min"`            //major and minor block device number
	Fstype     string        `json:"fstype"`             //filesystem type
	Mountpoint string        `json:"mountpoint"`         //block device mountpoint
	Label      string        `json:"label"`              //filesystem label
	Uuid       string        `json:"uuid"`               //filesystem uuid
	Parttype   string        `json:"parttype"`           //partition type UUID
	Partuuid   string        `json:"partuuid"`           //partition uuid
	Partflags  string        `json:"partflags"`          //partition flags
	Ra         string        `json:"ra"`                 //read-ahead of the device
	Ro         string        `json:"ro"`                 //read-only device
	Rm         string        `json:"rm"`                 //is device removable
	Hotplug    string        `json:"hotplug"`            //removable or hotplug device (usb, pcmcia, ...)
	Model      string        `json:"model"`              //device model number
	Serial     string        `json:"serial"`             //disk serial number
	Size       string        `json:"size"`               //size of device
	State      string        `json:"state"`              //state of the device
	Owner      string        `json:"owner"`              //owner user name
	Group      string        `json:"group"`              //group user name
	Mode       string        `json:"mode"`               //device node permissions
	Alignment  string        `json:"alignment"`          //alignment offset
	Minio      string        `json:"minio"`              //minimum I/O size
	Optio      string        `json:"optio"`              //optimal I/O size
	Physec     string        `json:"physec"`             //physical sector size
	Logsec     string        `json:"logsec"`             //logical sector size
	Rota       string        `json:"rota"`               //rotational device
	Sched      string        `json:"sched"`              //scheduler name
	Rqsize     string        `json:"rqsize"`             //request queue size
	Type       string        `json:"type"`               //is device disk or partition
	Discaln    string        `json:"discaln"`            //discard alignment offset
	Discgran   string        `json:"discgran"`           //discard granularity
	Discmax    string        `json:"discmax"`            //discard max bytes
	Disczero   string        `json:"disczero"`           //discard zeroes data
	Wsame      string        `json:"wsame"`              //write same max bytes
	Wwn        string        `json:"wwn"`                //unique storage identifier
	Rand       string        `json:"rand"`               //adds randomness
	Pkname     string        `json:"pkname"`             //internal parent kernel device name
	Hctl       string        `json:"hctl"`               //Host:Channel:Target:Lun for SCSI
	Tran       string        `json:"tran"`               //device transport type
	Subsystems string        `json:"subsystems"`         //de-duplicated chain of subsystems
	Rev        string        `json:"rev"`                //device revision
	Vendor     string        `json:"vendor"`             //device vendor
	Children   []Blockdevice `json:"children,omitempty"` //Blockdevice ...
}
