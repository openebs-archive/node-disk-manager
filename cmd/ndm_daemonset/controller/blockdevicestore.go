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

	"k8s.io/klog"
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
		klog.Info("eventcode=ndm.blockdevice.create.success",
			"msg=Created blockdevice object in etcd", "rname=", blockDeviceCopy.ObjectMeta.Name)
		return
	}

	if !errors.IsAlreadyExists(err) {
		klog.Error("eventcode=ndm.blockdevice.create.failure",
			"msg=Creation of blockdevice object failed: ", err,
			"rname=", blockDeviceCopy.ObjectMeta.Name)
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
		klog.Error("Updating of BlockDevice Object failed: ", err)
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
	klog.Error("Update to blockdevice object failed: ", blockDevice.ObjectMeta.Name)
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
			klog.Errorf("eventcode=%s msg=%s reason=Unable to get blockdevice object:%v, err:%v rname=%s",
				"ndm.blockdevice.update.failure", "Failed to update block device",
				oldBlockDevice.ObjectMeta.Name, err, blockDeviceCopy.ObjectMeta.Name)
			return err
		}
	}

	blockDeviceCopy.ObjectMeta.ResourceVersion = oldBlockDevice.ObjectMeta.ResourceVersion
	blockDeviceCopy.Spec.ClaimRef = oldBlockDevice.Spec.ClaimRef
	blockDeviceCopy.Status.ClaimState = oldBlockDevice.Status.ClaimState
	err = c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Error("eventcode=ndm.blockdevice.update.failure",
			"msg=Unable to update blockdevice object : ", err, "rname=", blockDeviceCopy.ObjectMeta.Name)
		return err
	}
	klog.Info("eventcode=ndm.blockdevice.update.success",
		"msg=Updated blockdevice object", "rname=", blockDeviceCopy.ObjectMeta.Name)
	return nil
}

// DeactivateBlockDevice API is used to set blockdevice status to "inactive" state in etcd
func (c *Controller) DeactivateBlockDevice(blockDevice apis.BlockDevice) {

	blockDeviceCopy := blockDevice.DeepCopy()
	blockDeviceCopy.Status.State = NDMInactive
	err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Error("eventcode=ndm.blockdevice.deactivate.failure",
			"msg=Unable to deactivate blockdevice: ", err, "rname=", blockDeviceCopy.ObjectMeta.Name)
		return
	}
	klog.Info("eventcode=ndm.blockdevice.deactivate.success",
		"msg=Deactivated blockdevice", "rname=", blockDeviceCopy.ObjectMeta.Name)
}

// GetBlockDevice get Disk resource from etcd
func (c *Controller) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	dvr := &apis.BlockDevice{}
	err := c.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, dvr)

	if err != nil {
		klog.Error("Unable to get blockdevice object : ", err)
		return nil, err
	}
	klog.Info("Got blockdevice object : ", name)
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
		klog.Error("eventcode=ndm.blockdevice.delete.failure",
			"msg=Unable to delete blockdevice object : ", err, "rname=", name)
		return
	}
	klog.Info("eventcode=ndm.blockdevice.delete.success",
		"msg=Deleted blockdevice object", "rname=", name)
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
		klog.Error(err)
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
		klog.Error(err)
		return
	}
	for _, item := range blockDeviceList.Items {
		blockDeviceCopy := item.DeepCopy()
		blockDeviceCopy.Status.State = NDMUnknown
		err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
		if err == nil {
			klog.Error("Status marked unknown for blockdevice object: ",
				blockDeviceCopy.ObjectMeta.Name)
		}
	}
}
