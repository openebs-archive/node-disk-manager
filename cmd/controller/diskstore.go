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

package controller

import (
	"github.com/golang/glog"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDisk creates the Disk resource in etcd
func (c *Controller) CreateDisk(dr apis.Disk) {
	drCopy := dr.DeepCopy()
	cdr, err := c.Clientset.OpenebsV1alpha1().Disks().Create(drCopy)
	if err == nil {
		glog.Info("created disk object : ", cdr.ObjectMeta.Name)
		return
	}
	/*
	 * creation failure can be due to the case that resource is already
	 * there with that uid. This is possible when disk has been moved from
	 * one node to another in a cluster. So here we just have to change the
	 * ownership of this disk resource to the current node.
	 */
	glog.Info("create failed, trying to update the disk object : ", err)
	err = c.UpdateDisk(dr, nil)
	if err == nil {
		return
	}
	/*
	 * updation failure can be due to the fact that old node may have set the status
	 * to Inactive after updateDr has done the Get call, as resource version will
	 * change with each update, so we have to try again. Also if other node tries to
	 * update the resource version after updation is successful here, the update call
	 * from that node will fail.
	 */
	glog.Info("disk status updated by other node, changing the ownership to this node : ", err)
	err = c.UpdateDisk(dr, nil)
	if err == nil {
		glog.Info("updated disk object in etcd : ", dr.ObjectMeta.Name)
		return
	}
}

// UpdateDisk update the Disk resource in etcd
func (c *Controller) UpdateDisk(dr apis.Disk, oldDr *apis.Disk) error {
	drCopy := dr.DeepCopy()
	if oldDr == nil {
		var err error
		oldDr, err = c.Clientset.OpenebsV1alpha1().Disks().Get(drCopy.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	drCopy.ObjectMeta.ResourceVersion = oldDr.ObjectMeta.ResourceVersion
	udr, err := c.Clientset.OpenebsV1alpha1().Disks().Update(drCopy)
	if err != nil {
		glog.Error("unable to update disk object : ", err)
		return err
	}
	glog.Info("updated disk object : ", udr.ObjectMeta.Name)
	return nil
}

// DeactivateDisk sets the disk status to inactive in etcd
func (c *Controller) DeactivateDisk(dr apis.Disk) {
	drCopy := dr.DeepCopy()
	drCopy.Status.State = NDMInactive
	udr, err := c.Clientset.OpenebsV1alpha1().Disks().Update(drCopy)
	if err != nil {
		glog.Error("unable to deactivate disk object : ", err)
		return
	}
	glog.Info("deactivate the disk object : ", udr.ObjectMeta.Name)
}

// DeleteDisk delete the Disk resource from etcd
func (c *Controller) DeleteDisk(name string) {
	err := c.Clientset.OpenebsV1alpha1().Disks().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		glog.Error("unable to delete disk object : ", err)
		return
	}
	glog.Info("deleted disk object : ", name)
}

// ListDiskResource queries the etcd for the devices for the host/node
// and returns list of disk resources.
func (c *Controller) ListDiskResource() (*apis.DiskList, error) {
	label := NDMHostKey + "=" + c.HostName
	filter := metav1.ListOptions{LabelSelector: label}
	listDR, err := c.Clientset.OpenebsV1alpha1().Disks().List(filter)
	return listDR, err
}

// GetExistingResource returns the existing disk resource if it is
// present in etcd if not it returns nil pointer.
func (c *Controller) GetExistingResource(listDr *apis.DiskList, uuid string) *apis.Disk {
	for _, item := range listDr.Items {
		if uuid == item.ObjectMeta.Name {
			return &item
		}
	}
	return nil
}

// DeactivateStaleDiskResource deactivates the stale entry from etcd.
// It gets list of resources which are present in system and queries etcd to get
// list of active resources. One active resource which is present in etcd not in
// system that will be marked as inactive.
func (c *Controller) DeactivateStaleDiskResource(devices []string) {
	listDR, err := c.ListDiskResource()
	if err != nil {
		glog.Error(err)
		return
	}
	for _, item := range listDR.Items {
		if !util.Contains(devices, item.ObjectMeta.Name) {
			c.DeactivateDisk(item)
		}
	}
}

// PushDiskResource is a utility function which checks old disk resource
// present or not. If it presents in etcd then it updates the resource
// else it creates one new disk resource in etcd
func (c *Controller) PushDiskResource(oldDr *apis.Disk, diskDetails *DiskInfo) {
	diskDetails.HostName = c.HostName
	diskDetails.Uuid = diskDetails.ProbeIdentifiers.Uuid
	diskApi := diskDetails.ToDisk()
	if oldDr != nil {
		c.UpdateDisk(diskApi, oldDr)
		return
	}
	c.CreateDisk(diskApi)
}
