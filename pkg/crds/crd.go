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

package crds

import apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

// CustomResource is a wrapper over the CustomResourceDefinition of apiextensions,
// used to build the CRD object
type CustomResource struct {
	object *apiext.CustomResourceDefinition
}

// GetAPIObject returns the CRD API from the wrapper struct
func (cr *CustomResource) GetAPIObject() *apiext.CustomResourceDefinition {
	return cr.object
}
