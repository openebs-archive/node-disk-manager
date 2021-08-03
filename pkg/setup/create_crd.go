/*
Copyright 2019 The OpenEBS Authors

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

// import (
// 	"encoding/json"
// 	"fmt"
// 	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
// 	"k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/apimachinery/pkg/util/wait"
// )

// // createBlockDeviceCRD creates a BlockDevice CRD
// func (sc Config) createBlockDeviceCRD() error {
// 	blockDeviceCRD, err := buildBlockDeviceCRD()
// 	if err != nil {
// 		return err
// 	}
// 	return sc.createCRD(blockDeviceCRD)
// }

// // createBlockDeviceClaimCRD creates a BlockDeviceClaim CRD
// func (sc Config) createBlockDeviceClaimCRD() error {
// 	blockDeviceClaimCRD, err := buildBlockDeviceClaimCRD()
// 	if err != nil {
// 		return err
// 	}
// 	return sc.createCRD(blockDeviceClaimCRD)
// }

// // createCRD creates a CRD in the cluster and waits for it to get into active state
// // It will return error, if the CRD creation failed, or the Name conflicts with other CRD already
// // in the group
// func (sc Config) createCRD(crd *apiext.CustomResourceDefinition) error {
// 	if _, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
// 		if errors.IsAlreadyExists(err) {
// 			// if CRD already exists, we patch it with the new changes.
// 			// This will also handle the upgrades of CRDs
// 			patch, err := json.Marshal(crd)
// 			if err != nil {
// 				return fmt.Errorf("could not marshal new customResourceDefintion for %s : %v", crd.Name, err)
// 			}
// 			if _, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Patch(crd.Name, types.MergePatchType, patch); err != nil {
// 				return fmt.Errorf("could not update customResourceDefinition for %s : %v", crd.Name, err)
// 			}
// 		} else {
// 			return err
// 		}
// 	}

// 	return wait.Poll(CRDRetryInterval, CRDTimeout, func() (done bool, err error) {
// 		c, err := sc.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
// 		if err != nil {
// 			return false, err
// 		}
// 		for _, cond := range c.Status.Conditions {
// 			switch cond.Type {
// 			case apiext.Established:
// 				if cond.Status == apiext.ConditionTrue {
// 					return true, err
// 				}
// 			case apiext.NamesAccepted:
// 				if cond.Status == apiext.ConditionFalse {
// 					return false, fmt.Errorf("name conflict for %s : %v", crd.Name, cond.Reason)
// 				}
// 			}
// 		}

// 		return false, err
// 	})
// }
