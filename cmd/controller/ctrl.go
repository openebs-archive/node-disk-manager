/*
Copyright 2018 OpenEBS Authors.

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
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/udev"
	"github.com/openebs/node-disk-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	apis "github.com/openebs/node-disk-manager/pkg/apis/ndm/v1alpha1"
	clientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned"
)

const (
	NDMKind     = "Disk"
	NDMVersion  = "openebs.io/v1alpha1"
	NDMPrefix   = "disk-"
	NDMHostKey  = "kubernetes.io/hostname"
	NDMActive   = "Active"
	NDMInactive = "Inactive"
)

// Controller is the controller implementation for do resources
type Controller struct {
	// node name
	hostname string
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	clientset clientset.Interface
}

// NewController returns a new sample controller
func NewController(
	host string,
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface) *Controller {

	controller := &Controller{
		hostname:      host,
		kubeclientset: kubeclientset,
		clientset:     clientset,
	}

	return controller
}

// getUid will return unique id for the disk block device
func getUid(device *udev.Udevice) string {
	return NDMPrefix +
		util.Hash(device.PropertyValue(udev.UDEV_WWN)+
			device.PropertyValue(udev.UDEV_MODEL)+
			device.PropertyValue(udev.UDEV_SERIAL)+
			device.PropertyValue(udev.UDEV_VENDOR))
}

// Run will discover the local disks and push them to the etcd
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	glog.Info("Started the controller")

	err := pushDiskResources(c)
	if err != nil {
		glog.Error("Unable to push disk objects to etcd : ", err)
		return err
	}
	<-stopCh
	glog.Info("Shutting down the controller")

	return nil
}

// DevList queries the etcd for the devices on the host
// and prints them to stdout
func (c *Controller) DevList() error {
	label := fmt.Sprintf("kubernetes.io/hostname=%v", c.hostname)

	filter := metav1.ListOptions{LabelSelector: label}
	listDR, err := c.clientset.NdmV1alpha1().Disks().List(filter)

	if err != nil {
		return err
	}

	for _, item := range listDR.Items {
		fmt.Printf("Path: %v\nSize: %v\nStatus: %v\nModel: %v\nSerial: %v\nVendor: %v\n\n",
			item.Spec.Path, item.Spec.Capacity.Storage, item.Status.State,
			item.Spec.Details.Model, item.Spec.Details.Serial, item.Spec.Details.Vendor)
	}
	return nil
}

// createDR creates the Disk resource in etcd
func createDR(dr apis.Disk, c *Controller) {
	drCopy := dr.DeepCopy()
	cdr, err := c.clientset.NdmV1alpha1().Disks().Create(drCopy)

	if err != nil {
		/*
		 * creation failure can be due to the case that resource is already
		 * there with that uid. This is possible when disk has been moved from
		 * one node to another in a cluster. So here we just have to change the
		 * ownership of this disk resource to the current node.
		 */
		glog.Info("create failed, trying to update the disk resource : ", err)
		err := updateDR(dr, nil, c)
		if err != nil {
			/*
			 * updation failure can be due to the fact that old node may have set the status
			 * to Inactive after updateDr has done the Get call, as resource version will
			 * change with each update, so we have to try again. Also if other node tries to
			 * update the resource version after updation is succssful here, the update call
			 * from that node will fail.
			 */
			glog.Info("disk status updated by other node, changing the ownership to this node : ", err)
			updateDR(dr, nil, c)
		}
	} else {
		glog.Info("Created disk object in etcd : ", cdr.ObjectMeta.Name)
	}
}

// updateDR update the Disk resource in etcd
func updateDR(dr apis.Disk, oldDr *apis.Disk, c *Controller) error {
	drCopy := dr.DeepCopy()
	if oldDr == nil {
		drGot, err := c.clientset.NdmV1alpha1().Disks().Get(drCopy.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			glog.Info("Unable to get the disk object from etcd : ", err)
			return err
		}
		drCopy.ObjectMeta.ResourceVersion = drGot.ObjectMeta.ResourceVersion
	} else {
		drCopy.ObjectMeta.ResourceVersion = oldDr.ObjectMeta.ResourceVersion
	}
	udr, err := c.clientset.NdmV1alpha1().Disks().Update(drCopy)

	if err != nil {
		glog.Info("Unable to update disk object to etcd : ", err)
		return err
	}

	glog.Info("Updated disk object to etcd : ", udr.ObjectMeta.Name)
	return nil
}

// deactivateDR sets the disk status to inactive in etcd
func deactivateDR(dr apis.Disk, c *Controller) {
	drCopy := dr.DeepCopy()
	drCopy.Status.State = NDMInactive
	udr, err := c.clientset.NdmV1alpha1().Disks().Update(drCopy)

	if err != nil {
		glog.Info("Unable to deactivate the disk object in etcd : ", err)
	} else {
		glog.Info("Deactivated the disk object : ", udr.ObjectMeta.Name)
	}
}

// deleteDR delete the Disk resource in etcd
func deleteDR(name string, c *Controller) {
	err := c.clientset.NdmV1alpha1().Disks().Delete(name, &metav1.DeleteOptions{})

	if err != nil {
		glog.Info("Unable to delete disk object from etcd : ", err)
	} else {
		glog.Info("Deleted disk object from etcd : ", name)
	}
}

// deactivateStaleDiskResource deactivates the stale entry from etcd
func deactivateStaleDiskResource(c *Controller, devices []*udev.Udevice, listDR *apis.DiskList) {
	for _, item := range listDR.Items {
		var uuid string
		for _, device := range devices {
			uuid = getUid(device)

			if uuid == item.ObjectMeta.Name {
				break
			}
		}
		if uuid != item.ObjectMeta.Name {
			deactivateDR(item, c)
		}
	}
}

// getExistingResource returns the existing disk resource
func getExistingResource(uuid string, listDR *apis.DiskList) *apis.Disk {
	for _, item := range listDR.Items {
		if uuid == item.ObjectMeta.Name {
			return &item
		}
	}
	return nil
}

// addNewDiskResource add the newly identified disks to the etcd
func addNewDiskResource(c *Controller, devices []*udev.Udevice, listDR *apis.DiskList) {
	for i := range devices {
		device := devices[i]
		/*
		 * filter partitions using Devtype and
		 * filter unwated devices (e.g CD ROM) using udev property
		 */
		if device.Devtype() == udev.UDEV_SYSTEM && device.PropertyValue(udev.UDEV_TYPE) == udev.UDEV_SYSTEM {
			var sector []byte
			var sec int64
			uuid := getUid(device)
			n, err := strconv.ParseInt(device.SysattrValue("size"), 10, 64)
			if err == nil {
				// should we use disk smart queries to get the sector size?
				fname := "/sys" + device.PropertyValue(udev.UDEV_PATH) + "/queue/hw_sector_size"
				sector, err = ioutil.ReadFile(fname)
				if err == nil {
					sec, err = strconv.ParseInt(string(sector[:len(sector)-1]), 10, 64)
				}
			}
			if err == nil {
				obj := apis.DiskSpec{Path: device.Devnode()}
				obj.Capacity.Storage = uint64(n * sec)
				obj.Details.Model = device.PropertyValue(udev.UDEV_MODEL)
				obj.Details.Serial = device.PropertyValue(udev.UDEV_SERIAL)
				obj.Details.Vendor = device.PropertyValue(udev.UDEV_VENDOR)

				dr := apis.Disk{Spec: obj}
				dr.Status.State = NDMActive
				dr.ObjectMeta.Name = uuid
				dr.ObjectMeta.Labels = make(map[string]string)
				dr.ObjectMeta.Labels[NDMHostKey] = c.hostname
				dr.TypeMeta.Kind = NDMKind
				dr.TypeMeta.APIVersion = NDMVersion
				edr := getExistingResource(uuid, listDR)
				if edr != nil {
					glog.Info("disk object already exist in etcd : ", uuid)
					//TODO: update only if disk properties have changed
					updateDR(dr, edr, c)
				} else {
					createDR(dr, c)
				}
			} else {
				glog.Info("error pushing disk object to etcd : ", err)
			}
		}
	}
}

// pushDiskResources push new disks and deactivates the stale entries
func pushDiskResources(c *Controller) error {
	glog.Info("pushDiskResources pushing disk object to etcd")

	udevices := udev.ListDevices()
	if udevices == nil {
		return errors.New("error listing attached disk")
	}

	label := fmt.Sprintf("kubernetes.io/hostname=%v", c.hostname)

	filter := metav1.ListOptions{LabelSelector: label}
	listDR, err := c.clientset.NdmV1alpha1().Disks().List(filter)

	if err != nil {
		return err
	}

	deactivateStaleDiskResource(c, udevices, listDR)

	addNewDiskResource(c, udevices, listDR)

	glog.Info("pushDiskResources done pushing disk object to etcd")

	return nil
}
