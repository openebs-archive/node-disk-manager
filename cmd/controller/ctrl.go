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
	"github.com/openebs/node-disk-manager/cmd/storage/block"
	"github.com/openebs/node-disk-manager/cmd/types/v1"
	"strconv"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	apis "github.com/openebs/node-disk-manager/pkg/apis/ndm/v1alpha1"
	clientset "github.com/openebs/node-disk-manager/pkg/client/clientset/versioned"
)

const (
	NDMKind    = "Disk"
	NDMVersion = "openebs.io/v1alpha1"
	NDMPrefix  = "disk-"
)

// Controller is the controller implementation for do resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	clientset clientset.Interface
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface) *Controller {

	controller := &Controller{
		kubeclientset: kubeclientset,
		clientset:     clientset,
	}

	return controller
}

// getUid will return unique id for the disk block device
func getUid(blk v1.Blockdevice) string {
	return NDMPrefix + util.Hash(blk.Wwn+blk.Model+blk.Serial+blk.Vendor)
}

// Run will discover the local disks and push them to the etcd
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	glog.Info("Started the controller")

	err := pushDiskResources(c)
	if err != nil {
		glog.Error("Unable to push disk objects to etcd : ", err)
	}
	<-stopCh
	glog.Info("Shutting down the controller")

	return nil
}

// createDR creates the Disk resource in etcd
func createDR(dr apis.Disk, c *Controller) {
	drCopy := dr.DeepCopy()
	cdr, err := c.clientset.NdmV1alpha1().Disks().Create(drCopy)

	if err != nil {
		glog.Info("Unable to create disk object in etcd : ", err)
	} else {
		glog.Info("Created disk object in etcd : ", cdr.ObjectMeta.Name)
	}
}

// updateDR update the Disk resource in etcd
func updateDR(dr apis.Disk, c *Controller) {
	drCopy := dr.DeepCopy()
	drGot, err := c.clientset.NdmV1alpha1().Disks().Get(drCopy.ObjectMeta.Name, metav1.GetOptions{})
	drGot.Spec = drCopy.Spec
	udr, err := c.clientset.NdmV1alpha1().Disks().Update(drGot)

	if err != nil {
		glog.Info("Unable to update disk object to etcd : ", err)
	} else {
		glog.Info("Updated disk object to etcd : ", udr.ObjectMeta.Name)
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

// deleteStaleDiskResource deletes the stale entry from etcd
func deleteStaleDiskResource(c *Controller, detected v1.BlockDeviceInfo, listDR *apis.DiskList) {
	for _, item := range listDR.Items {
		var uuid string
		for _, blk := range detected.Blockdevices {
			uuid = getUid(blk)

			if uuid == item.ObjectMeta.Name {
				break
			}
		}
		if uuid != item.ObjectMeta.Name {
			deleteDR(item.ObjectMeta.Name, c)
		}
	}
}

// isDiskReousrceExist checks if disk resource exist in etcd
func isDiskReousrceExist(uuid string, listDR *apis.DiskList) bool {
	for _, item := range listDR.Items {
		if uuid == item.ObjectMeta.Name {
			return true
		}
	}
	return false
}

// addNewDiskResource add the newly identified disks to the etcd
func addNewDiskResource(c *Controller, detected v1.BlockDeviceInfo, listDR *apis.DiskList) {
	for _, blk := range detected.Blockdevices {
		if blk.Type == "disk" {
			uuid := getUid(blk)

			if isDiskReousrceExist(uuid, listDR) {
				glog.Info("disk object already exist in etcd : ", blk.Name)
			} else {
				glog.Info("pushing disk object to etcd : ", blk.Name)
				obj := apis.DiskSpec{Path: "/dev/" + blk.Name}
				n, err := strconv.ParseInt(blk.Size, 10, 64)
				if err == nil {
					obj.Capacity.Storage = uint64(n)
					obj.Details.Model = blk.Model
					obj.Details.Serial = blk.Serial
					obj.Details.Vendor = blk.Vendor

					dr := apis.Disk{Spec: obj}
					dr.ObjectMeta.Name = uuid
					dr.TypeMeta.Kind = NDMKind
					dr.TypeMeta.APIVersion = NDMVersion
					createDR(dr, c)
				} else {
					glog.Info("error pushing disk object to etcd : ", err)
				}
			}
		}
	}
}

// pushDiskResources push new disks and deletes stale entries form etcd
func pushDiskResources(c *Controller) error {
	var resJsonDecoded v1.BlockDeviceInfo
	err := block.ListBlockExec(&resJsonDecoded)
	if err != nil {
		return err
	}

	listDR, err := c.clientset.NdmV1alpha1().Disks().List(metav1.ListOptions{})

	if err != nil {
		return err
	}

	deleteStaleDiskResource(c, resJsonDecoded, listDR)

	addNewDiskResource(c, resJsonDecoded, listDR)

	return nil
}
