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
	"strings"

	"github.com/golang/glog"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cntrlutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// API to creates the NdmConfig resource in etcd
// This API will be called once by each ndm-daemonset pod
func (c *Controller) CreateNdmConfig(ndmconfigR *apis.NdmConfig) {

	ndmconfigRCopy := ndmconfigR.DeepCopy()
	err := c.Clientset.Create(context.TODO(), ndmconfigRCopy)
	if err == nil {
		glog.Info("Created ndmconfig object in etcd: ",
			ndmconfigRCopy.ObjectMeta.Name)
		return
	}

	if !errors.IsAlreadyExists(err) {
		glog.Error("Creation of ndmconfig object failed: ", err)
		return
	}

	glog.Error("Update to device object failed: ", ndmconfigR.ObjectMeta.Name)
}

// GetNDMconfigR get NdmConfig resource from etcd
func (c *Controller) GetGetNDMconfigR(name string) (*apis.NdmConfig, error) {
	ndmconfigR := &apis.NdmConfig{}
	err := c.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: "", Name: name}, ndmconfigR)

	if err != nil {
		glog.Error("Unable to get NdmConfig object : ", err)
		return nil, err
	}
	glog.Info("Got NdmConfig object : ", name)
	return ndmconfigR, nil
}

// ListNdmConfigResource queries the etcd for the NdmConfig for the host/node
// and returns NdmConfig resource
func (c *Controller) ListNdmConfigResource() (*apis.NdmConfigList, error) {

	listNdmConfigR := &apis.NdmConfigList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NdmConfig",
			APIVersion: "openebs.io/v1alpha1",
		},
	}

	filter := NDMHostKey + "=" + c.HostName
	opts := &client.ListOptions{}
	opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, listNdmConfigR)
	return listNdmConfigR, err
}

func (C *Controller) getObjectMeta(name string, ns string) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Labels:    make(map[string]string),
		Name:      name,
		Namespace: ns,
	}
	objectMeta.Labels[NDMHostKey] = C.HostName
	return objectMeta
}

func (C *Controller) getTypeMeta() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       "NdmConfig",
		APIVersion: "openebs.io/v1alpha1",
	}
	return typeMeta
}

func (C *Controller) getStatus() apis.NdmConfigStatus {
	Status := apis.NdmConfigStatus{
		Phase: apis.NdmConfigPhaseInit,
	}
	return Status
}

// Push NdmConfig-CR per ndm-daemonset pod
// Pod name will be used as name of NdmConfig-CR
// Hostname will be added into label
func (c *Controller) PushNdmConfigResource() error {

	// Get list of ndm pods
	podList := &corev1.PodList{}
	opts := &client.ListOptions{}
	filter := "name" + "=" + "node-disk-manager"
	glog.Info("Filter string", "filter:", filter)

	opts.SetLabelSelector(filter)
	err := c.Clientset.List(context.TODO(), opts, podList)
	if err != nil {
		glog.Info("Error in getting ndm pod list")
		return err
	}

	if len(podList.Items) == 0 {
		glog.Info("podList is nil")
	} else {
		glog.Info("podList is non-nil")
	}

	for _, pod := range podList.Items {
		glog.Info("Pod:", pod.ObjectMeta.Name)
		if strings.Contains(pod.Spec.NodeName, c.HostName) {
			glog.Infof("Found corresponding pod:%#v:", pod)
			ndmConfigR := &apis.NdmConfig{}
			Rname := NDMConfigPreFix + pod.ObjectMeta.Name
			ndmConfigR.ObjectMeta = c.getObjectMeta(Rname, pod.ObjectMeta.Namespace)
			ndmConfigR.TypeMeta = c.getTypeMeta()
			ndmConfigR.Status = c.getStatus()
			// Set Memcached instance as the owner and controller
			err = cntrlutils.SetControllerReference(&pod, ndmConfigR, c.mgr.GetScheme())
			if err != nil {
				return err
			}
			c.CreateNdmConfig(ndmConfigR)
			return nil
		}
	}
	return nil
}
