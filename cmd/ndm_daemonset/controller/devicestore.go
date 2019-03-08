/*
Copyright 2019 The OpenEBS Authors.

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

package controller

import (
	"context"

	"github.com/golang/glog"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// API to creates the Device resource in etcd
// This API will be called for each new addDiskEvent
// dvr is DeviceResource-CR
func (c *Controller) CreateDevice(dvr apis.Device) {

	dvrCopy := dvr.DeepCopy()
	err := c.Clientset.Create(context.TODO(), dvrCopy)
	if err == nil {
		glog.Info("Created device object in etcd: ",
			dvrCopy.ObjectMeta.Name)
		return
	}

	if !errors.IsAlreadyExists(err) {
		glog.Error("Creation of device object failed: ", err)
		return
	}

	/*
	 * Creation may fail because resource is already exist in etcd.
	 * This is possible when disk moved from one node to another in
	 * cluster so device object need to be updated with new Node.
	 */
	err = c.UpdateDevice(dvr, nil)
	if err == nil {
		return
	}

	if !errors.IsConflict(err) {
		glog.Error("Updating of Device Object failed: ", err)
		return
	}

	/*
	 * Update might failed due to to resource version mismatch which
	 * can happen if some other entity updating same resource in parallel.
	 */
	err = c.UpdateDevice(dvr, nil)
	if err == nil {
		return
	}
	glog.Error("Update to device object failed: ", dvr.ObjectMeta.Name)
}

// UpdateDevice update the Device resource in etcd
func (c *Controller) UpdateDevice(dvr apis.Device, oldDvr *apis.Device) error {
	var err error

	dvrCopy := dvr.DeepCopy()
	if oldDvr == nil {
		oldDvrCopy := dvr.DeepCopy()
		err = c.Clientset.Get(context.TODO(), client.ObjectKey{
			Namespace: oldDvrCopy.Namespace,
			Name:      oldDvrCopy.Name}, oldDvrCopy)
		if err != nil {
			glog.Error("Unable to get device object : ", err)
			return err
		}
		dvrCopy.ObjectMeta.ResourceVersion = oldDvrCopy.ObjectMeta.ResourceVersion
	}

	err = c.Clientset.Update(context.TODO(), dvrCopy)
	if err != nil {
		glog.Error("Unable to update device object : ", err)
		return err
	}
	glog.Info("Updated device object : ", dvrCopy.ObjectMeta.Name)
	return nil
}

// This API used to set device status to "inactive" state in etcd
func (c *Controller) DeactivateDevice(dvr apis.Device) {

	dvrCopy := dvr.DeepCopy()
	dvrCopy.Status.State = NDMInactive
	err := c.Clientset.Update(context.TODO(), dvrCopy)
	if err != nil {
		glog.Error("Unable to deactivate device: ", err)
		return
	}
	glog.Info("Deactivated device: ", dvrCopy.ObjectMeta.Name)
}

// GetDisk get Disk resource from etcd
func (c *Controller) GetDevice(name string) (*apis.Device, error) {
	dvr := &apis.Device{}
	err := c.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, dvr)

	if err != nil {
		glog.Error("Unable to get device object : ", err)
		return nil, err
	}
	glog.Info("Got device object : ", name)
	return dvr, nil
}

// DeleteDevice delete the Device resource from etcd
func (c *Controller) DeleteDevice(name string) {
	dvr := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   name,
		},
	}

	err := c.Clientset.Delete(context.TODO(), dvr)
	if err != nil {
		glog.Error("Unable to delete device object : ", err)
		return
	}
	glog.Info("Deleted device object : ", name)
}

// ListDeviceResource queries the etcd for the devices for the host/node
// and returns list of device resources.
func (c *Controller) ListDeviceResource() (*apis.DeviceList, error) {

	listDVR := &apis.DeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	filter := NDMHostKey + "=" + c.HostName
	opts := &client.ListOptions{}
	opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, listDVR)
	return listDVR, err
}

// GetExistingDeviceResource returns the existing device resource if it is
// present in etcd if not it returns nil pointer.
func (c *Controller) GetExistingDeviceResource(listDvr *apis.DeviceList,
	uuid string) *apis.Device {
	for _, item := range listDvr.Items {
		if uuid == item.ObjectMeta.Name {
			return &item
		}
	}
	return nil
}

// DeactivateStaleDeviceResource deactivates the stale entry from etcd.
// It gets list of resources which are present in system and queries etcd to get
// list of active resources. Active resource which is present in etcd not in
// system that will be marked as inactive.
func (c *Controller) DeactivateStaleDeviceResource(devices []string) {
	listDevices := append(devices, GetActiveSparseDisksUuids(c.HostName)...)
	listDVR, err := c.ListDeviceResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range listDVR.Items {
		if !util.Contains(listDevices, item.ObjectMeta.Name) {
			c.DeactivateDevice(item)
		}
	}
}

// PushDeviceResource is a utility function which checks old device resource
// present or not. If it presents in etcd then it updates the resource
// else it creates new device resource in etcd
func (c *Controller) PushDeviceResource(oldDvr *apis.Device,
	deviceDetails *DeviceInfo) {
	deviceDetails.HostName = c.HostName
	deviceApi := deviceDetails.ToDevice()
	if oldDvr != nil {
		c.UpdateDevice(deviceApi, oldDvr)
		return
	}
	c.CreateDevice(deviceApi)
}

// MarkDeviceStatusToUnknown makes state of all resources owned by node unknown
// This will call as a cleanup process before shutting down.
func (c *Controller) MarkDeviceStatusToUnknown() {
	listDVR, err := c.ListDeviceResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range listDVR.Items {
		dvrCopy := item.DeepCopy()
		dvrCopy.Status.State = NDMUnknown
		err := c.Clientset.Update(context.TODO(), dvrCopy)
		if err == nil {
			glog.Error("Status marked unknown for device object: ",
				dvrCopy.ObjectMeta.Name)
		}
	}
}
