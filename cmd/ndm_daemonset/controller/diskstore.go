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
	"context"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateDisk creates the Disk resource in etcd
func (c *Controller) CreateDisk(dr apis.Disk) {
	drCopy := dr.DeepCopy()
	err := c.Clientset.Create(context.TODO(), drCopy)
	if err == nil {
		klog.Infof("eventcode=%s msg=%s rname=%v",
			"ndm.disk.create.success", "Created disk object in etcd",
			drCopy.ObjectMeta.Name)
		return
	}
	/*
	 * creation failure can be due to the case that resource is already
	 * there with that uid. This is possible when disk has been moved from
	 * one node to another in a cluster. So here we just have to change the
	 * ownership of this disk resource to the current node.
	 */
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
	klog.Info("disk status updated by other node, changing the ownership to this node : ", err)
	err = c.UpdateDisk(dr, nil)
	if err == nil {
		klog.Info("updated disk object in etcd : ", dr.ObjectMeta.Name)
		return
	}
}

// UpdateDisk update the Disk resource in etcd
func (c *Controller) UpdateDisk(dr apis.Disk, oldDr *apis.Disk) error {
	drCopy := dr.DeepCopy()

	if oldDr == nil {
		oldDr = dr.DeepCopy()
		var err error
		err = c.Clientset.Get(context.TODO(), client.ObjectKey{
			Namespace: oldDr.Namespace,
			Name:      oldDr.Name}, oldDr)
		if err != nil {
			klog.Errorf("eventcode=%s msg=%s : %v, err:%v rname=%v",
				"ndm.disk.update.failure", "Unable to get disk object",
				oldDr.ObjectMeta.Name, err, drCopy.ObjectMeta.Name)
			return err
		}
	}

	drCopy.ObjectMeta.ResourceVersion = oldDr.ObjectMeta.ResourceVersion
	err := c.Clientset.Update(context.TODO(), drCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s, err:%v rname=%v",
			"ndm.disk.update.failure", "Unable to update disk object",
			err, drCopy.ObjectMeta.Name)
		return err
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.disk.update.success", "Updated disk object successfully",
		drCopy.ObjectMeta.Name)
	return nil
}

// DeactivateDisk sets the disk status to inactive in etcd
func (c *Controller) DeactivateDisk(dr apis.Disk) {
	drCopy := dr.DeepCopy()
	drCopy.Status.State = NDMInactive
	err := c.Clientset.Update(context.TODO(), drCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.disk.deactivate.failure", "Unable to deactivate disk object",
			err, drCopy.ObjectMeta.Name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.disk.deactivate.success", "Deactivated the disk object",
		drCopy.ObjectMeta.Name)
}

// GetDisk get Disk resource from etcd
func (c *Controller) GetDisk(name string) (*apis.Disk, error) {
	dr := &apis.Disk{}
	err := c.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, dr)

	if err != nil {
		klog.Error("Unable to get disk object : ", err)
		return nil, err
	}
	klog.Info("Got disk object : ", name)
	return dr, nil
}

// DeleteDisk delete the Disk resource from etcd
func (c *Controller) DeleteDisk(name string) {
	dr := &apis.Disk{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   name,
		},
	}

	err := c.Clientset.Delete(context.TODO(), dr)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.disk.delete.failure", "Unable to delete disk object",
			err, name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.disk.delete.success", "Deleted disk object",
		name)
}

// ListDiskResource queries the etcd for the devices for the host/node
// and returns list of disk resources.
func (c *Controller) ListDiskResource() (*apis.DiskList, error) {
	listDR := &apis.DiskList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Disk",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	filter := KubernetesHostNameLabel + "=" + c.NodeAttributes[HostNameKey]
	filter = filter + "," + NDMManagedKey + "!=" + FalseString
	opts := &client.ListOptions{}
	opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, listDR)
	return listDR, err
}

// GetExistingDiskResource returns the existing disk resource if it is
// present in etcd if not it returns nil pointer.
func (c *Controller) GetExistingDiskResource(listDr *apis.DiskList, uuid string) *apis.Disk {
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
func (c *Controller) DeactivateStaleDiskResource(disks []string) {
	listDR, err := c.ListDiskResource()
	if err != nil {
		klog.Error(err)
		return
	}
	for _, item := range listDR.Items {
		if !util.Contains(disks, item.ObjectMeta.Name) {
			c.DeactivateDisk(item)
		}
	}
}

// PushDiskResource is a utility function which checks old disk resource
// present or not. If it presents in etcd then it updates the resource
// else it creates one new disk resource in etcd
func (c *Controller) PushDiskResource(oldDr *apis.Disk, diskDetails *DiskInfo) {
	diskDetails.Uuid = diskDetails.ProbeIdentifiers.Uuid
	diskDetails.NodeAttributes = c.NodeAttributes
	diskApi := diskDetails.ToDisk()
	if oldDr != nil {
		c.UpdateDisk(diskApi, oldDr)
		return
	}
	c.CreateDisk(diskApi)
}

// MarkDiskStatusToUnknown makes state of all resources owned by node unknown
// This will call as a cleanup process before shutting down.
func (c *Controller) MarkDiskStatusToUnknown() {
	listDR, err := c.ListDiskResource()
	if err != nil {
		klog.Error(err)
		return
	}
	for _, item := range listDR.Items {
		drCopy := item.DeepCopy()
		drCopy.Status.State = NDMUnknown
		err := c.Clientset.Update(context.TODO(), drCopy)
		if err == nil {
			klog.Error("updated disk object : ", drCopy.ObjectMeta.Name)
		}
	}
}
