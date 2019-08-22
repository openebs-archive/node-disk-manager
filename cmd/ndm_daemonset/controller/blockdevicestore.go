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

// CreateBlockDevice creates the BlockDevice resource in etcd
// This API will be called for each new addDiskEvent
// blockDevice is DeviceResource-CR
func (c *Controller) CreateBlockDevice(blockDevice apis.BlockDevice) {

	blockDeviceCopy := blockDevice.DeepCopy()
	err := c.Clientset.Create(context.TODO(), blockDeviceCopy)
	if err == nil {
		glog.Info("Created blockdevice object in etcd: ",
			blockDeviceCopy.ObjectMeta.Name)
		return
	}

	if !errors.IsAlreadyExists(err) {
		glog.Error("Creation of blockdevice object failed: ", err)
		return
	}

	/*
	 * Creation may fail because resource is already exist in etcd.
	 * This is possible when disk moved from one node to another in
	 * cluster so blockdevice object need to be updated with new Node.
	 */
	err = c.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return
	}

	if !errors.IsConflict(err) {
		glog.Error("Updating of BlockDevice Object failed: ", err)
		return
	}

	/*
	 * Update might failed due to to resource version mismatch which
	 * can happen if some other entity updating same resource in parallel.
	 */
	err = c.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return
	}
	glog.Error("Update to blockdevice object failed: ", blockDevice.ObjectMeta.Name)
}

// UpdateBlockDevice update the BlockDevice resource in etcd
func (c *Controller) UpdateBlockDevice(blockDevice apis.BlockDevice, oldBlockDevice *apis.BlockDevice) error {
	var err error

	blockDeviceCopy := blockDevice.DeepCopy()
	if oldBlockDevice == nil {
		oldBlockDevice = blockDevice.DeepCopy()
		err = c.Clientset.Get(context.TODO(), client.ObjectKey{
			Namespace: oldBlockDevice.Namespace,
			Name:      oldBlockDevice.Name}, oldBlockDevice)
		if err != nil {
			glog.Errorf("Unable to get blockdevice object:%v, err:%v", oldBlockDevice.ObjectMeta.Name, err)
			return err
		}
	}

	blockDeviceCopy.ObjectMeta.ResourceVersion = oldBlockDevice.ObjectMeta.ResourceVersion
	blockDeviceCopy.Spec.ClaimRef = oldBlockDevice.Spec.ClaimRef
	blockDeviceCopy.Status.ClaimState = oldBlockDevice.Status.ClaimState
	err = c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		glog.Error("Unable to update blockdevice object : ", err)
		return err
	}
	glog.Info("Updated blockdevice object : ", blockDeviceCopy.ObjectMeta.Name)
	return nil
}

// DeactivateBlockDevice API is used to set blockdevice status to "inactive" state in etcd
func (c *Controller) DeactivateBlockDevice(blockDevice apis.BlockDevice) {

	blockDeviceCopy := blockDevice.DeepCopy()
	blockDeviceCopy.Status.State = NDMInactive
	err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		glog.Error("Unable to deactivate blockdevice: ", err)
		return
	}
	glog.Info("Deactivated blockdevice: ", blockDeviceCopy.ObjectMeta.Name)
}

// GetBlockDevice get Disk resource from etcd
func (c *Controller) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	dvr := &apis.BlockDevice{}
	err := c.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, dvr)

	if err != nil {
		glog.Error("Unable to get blockdevice object : ", err)
		return nil, err
	}
	glog.Info("Got blockdevice object : ", name)
	return dvr, nil
}

// DeleteBlockDevice delete the BlockDevice resource from etcd
func (c *Controller) DeleteBlockDevice(name string) {
	blockDevice := &apis.BlockDevice{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   name,
		},
	}

	err := c.Clientset.Delete(context.TODO(), blockDevice)
	if err != nil {
		glog.Error("Unable to delete blockdevice object : ", err)
		return
	}
	glog.Info("Deleted blockdevice object : ", name)
}

// ListBlockDeviceResource queries the etcd for the devices for the host/node
// and returns list of blockdevice resources.
func (c *Controller) ListBlockDeviceResource() (*apis.BlockDeviceList, error) {

	blockDeviceList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	filter := KubernetesHostNameLabel + "=" + c.NodeAttributes[HostNameKey]
	filter = filter + "," + NDMManagedKey + "!=" + FalseString
	opts := &client.ListOptions{}
	opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, blockDeviceList)
	return blockDeviceList, err
}

// GetExistingBlockDeviceResource returns the existing blockdevice resource if it is
// present in etcd if not it returns nil pointer.
func (c *Controller) GetExistingBlockDeviceResource(blockDeviceList *apis.BlockDeviceList,
	uuid string) *apis.BlockDevice {
	for _, item := range blockDeviceList.Items {
		if uuid == item.ObjectMeta.Name {
			return &item
		}
	}
	return nil
}

// DeactivateStaleBlockDeviceResource deactivates the stale entry from etcd.
// It gets list of resources which are present in system and queries etcd to get
// list of active resources. Active resource which is present in etcd not in
// system that will be marked as inactive.
func (c *Controller) DeactivateStaleBlockDeviceResource(devices []string) {
	listDevices := append(devices, GetActiveSparseBlockDevicesUUID(c.NodeAttributes[HostNameKey])...)
	blockDeviceList, err := c.ListBlockDeviceResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range blockDeviceList.Items {
		if !util.Contains(listDevices, item.ObjectMeta.Name) {
			c.DeactivateBlockDevice(item)
		}
	}
}

// PushBlockDeviceResource is a utility function which checks old blockdevice resource
// present or not. If it presents in etcd then it updates the resource
// else it creates new blockdevice resource in etcd
func (c *Controller) PushBlockDeviceResource(oldBlockDevice *apis.BlockDevice,
	deviceDetails *DeviceInfo) {
	deviceDetails.NodeAttributes = c.NodeAttributes
	deviceAPI := deviceDetails.ToDevice()
	if oldBlockDevice != nil {
		c.UpdateBlockDevice(deviceAPI, oldBlockDevice)
		return
	}
	c.CreateBlockDevice(deviceAPI)
}

// MarkBlockDeviceStatusToUnknown makes state of all resources owned by node unknown
// This will call as a cleanup process before shutting down.
func (c *Controller) MarkBlockDeviceStatusToUnknown() {
	blockDeviceList, err := c.ListBlockDeviceResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range blockDeviceList.Items {
		blockDeviceCopy := item.DeepCopy()
		blockDeviceCopy.Status.State = NDMUnknown
		err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
		if err == nil {
			glog.Error("Status marked unknown for blockdevice object: ",
				blockDeviceCopy.ObjectMeta.Name)
		}
	}
}
