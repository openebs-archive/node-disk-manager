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
