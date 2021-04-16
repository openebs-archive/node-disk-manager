/*
Copyright 2020 The OpenEBS Authors

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

package setup

import (
	"github.com/openebs/node-disk-manager/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (sc Config) deleteDiskCRD() error {
	diskCRDName := "disks" + "." + v1alpha1.GroupName
	return sc.deleteCRD(diskCRDName)
}

func (sc Config) deleteCRD(crdName string) error {
	// check if crd exists,
	_, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, v1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// if exists delete the crd
	propagationPolicy := v1.DeletePropagationForeground
	deleteOptions := &v1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
	if err = sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().
		Delete(crdName, deleteOptions); err != nil {
		return err
	}
	return err
}
