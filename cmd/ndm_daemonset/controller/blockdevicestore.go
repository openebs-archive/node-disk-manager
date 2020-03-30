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

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateBlockDevice creates the BlockDevice resource in etcd
// This API will be called for each new addDiskEvent
// blockDevice is DeviceResource-CR
func (c *Controller) CreateBlockDevice(blockDevice apis.BlockDevice) error {

	blockDeviceCopy := blockDevice.DeepCopy()
	err := c.Clientset.Create(context.TODO(), blockDeviceCopy)
	if err == nil {
		klog.Infof("eventcode=%s msg=%s rname=%v",
			"ndm.blockdevice.create.success", "Created blockdevice object in etcd",
			blockDeviceCopy.ObjectMeta.Name)
		return err
	}

	if !errors.IsAlreadyExists(err) {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.create.failure", "Creation of blockdevice object failed",
			err, blockDeviceCopy.ObjectMeta.Name)
		return err
	}

	/*
	 * Creation may fail because resource is already exist in etcd.
	 * This is possible when disk moved from one node to another in
	 * cluster so blockdevice object need to be updated with new Node.
	 */
	err = c.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return err
	}

	if !errors.IsConflict(err) {
		klog.Error("Updating of BlockDevice Object failed: ", err)
		return err
	}

	/*
	 * Update might failed due to to resource version mismatch which
	 * can happen if some other entity updating same resource in parallel.
	 */
	err = c.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return err
	}
	klog.Error("Update to blockdevice object failed: ", blockDevice.ObjectMeta.Name)
	return nil
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
			klog.Errorf("eventcode=%s msg=%s : %v, err:%v rname=%v",
				"ndm.blockdevice.update.failure",
				"Failed to update block device : unable to get blockdevice object",
				oldBlockDevice.ObjectMeta.Name, err, blockDeviceCopy.ObjectMeta.Name)
			return err
		}
	}

	blockDeviceCopy.ObjectMeta = mergeMetadata(blockDeviceCopy.ObjectMeta, oldBlockDevice.ObjectMeta)
	blockDeviceCopy.Spec.ClaimRef = oldBlockDevice.Spec.ClaimRef
	blockDeviceCopy.Status.ClaimState = oldBlockDevice.Status.ClaimState
	err = c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.update.failure", "Unable to update blockdevice object",
			err, blockDeviceCopy.ObjectMeta.Name)
		return err
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.update.success", "Updated blockdevice object",
		blockDeviceCopy.ObjectMeta.Name)
	return nil
}

// DeactivateBlockDevice API is used to set blockdevice status to "inactive" state in etcd
func (c *Controller) DeactivateBlockDevice(blockDevice apis.BlockDevice) {

	blockDeviceCopy := blockDevice.DeepCopy()
	blockDeviceCopy.Status.State = NDMInactive
	err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v ",
			"ndm.blockdevice.deactivate.failure", "Unable to deactivate blockdevice",
			err, blockDeviceCopy.ObjectMeta.Name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.deactivate.success", "Deactivated blockdevice",
		blockDeviceCopy.ObjectMeta.Name)
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
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.delete.failure", "Unable to delete blockdevice object",
			err, name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.delete.success", "Deleted blockdevice object ", name)
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
	_ = opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, blockDeviceList)
	if err != nil {
		return blockDeviceList, err
	}

	// applying annotation filter, so that blockdevice resources that need not be reconciled are
	// not updated by the daemon
	for i := 0; i < len(blockDeviceList.Items); i++ {
		// if the annotation exists and the value is false, then that blockdevice resource will be removed
		// from the list
		if val, ok := blockDeviceList.Items[i].Annotations[OpenEBSReconcile]; ok && util.CheckFalsy(val) {
			blockDeviceList.Items = append(blockDeviceList.Items[:i], blockDeviceList.Items[i+1:]...)
		}
	}
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
	deviceDetails *DeviceInfo) error {
	deviceDetails.NodeAttributes = c.NodeAttributes
	deviceAPI := deviceDetails.ToDevice()
	if oldBlockDevice != nil {
		return c.UpdateBlockDevice(deviceAPI, oldBlockDevice)
	}
	return c.CreateBlockDevice(deviceAPI)
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

// mergeMetadata merges oldMetadata with newMetadata. It takes old metadata and
// update it's value with the help of new metadata.
func mergeMetadata(newMetadata, oldMetadata metav1.ObjectMeta) metav1.ObjectMeta {
	// metadata of older object which contains -
	// - name - no patch required we can use old object.
	// - namespace - no patch required we can use old object.
	// - generateName - no patch required we are not using it.
	// - selfLink - populated by the system we should use old object.
	// - uid - populated by the system we should use old object.
	// - resourceVersion - populated by the system we should use old object.
	// - generation - populated by the system we should use old object.
	// - creationTimestamp - populated by the system we should use old object.
	// - deletionTimestamp - populated by the system we should use old object.
	// - deletionGracePeriodSeconds - populated by the system we should use old object.
	// - labels - we will patch older labels with new labels.
	// - annotations - we will patch older annotations with new annotations.
	// - ownerReferences as ndm-ds is not adding ownerReferences we can go with old object.
	// - initializers ^^^
	// - finalizers ^^^
	// - clusterName - no patch required we can use old object.

	// Patch older label with new label. If there is a new key then it will be added
	// if it is an existing key then value will be overwritten with value from new label
	for key, value := range newMetadata.Labels {
		oldMetadata.Labels[key] = value
	}

	// Patch older annotations with new annotations. If there is a new key then it will be added
	// if it is an existing key then value will be overwritten with value from new annotations
	for key, value := range newMetadata.Annotations {
		oldMetadata.Annotations[key] = value
	}

	return oldMetadata
}
