package crds

import apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

type CustomResource struct {
	object *apiext.CustomResourceDefinition
}

func (cr *CustomResource) GetAPIObject() *apiext.CustomResourceDefinition {
	return cr.object
}
