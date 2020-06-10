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

import (
	"errors"
	"fmt"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// Builder is struct used to build the CRD object
type Builder struct {
	crd  *CustomResource
	errs []error
}

// NewBuilder returns a new builder for creating CRD
func NewBuilder() *Builder {
	return &Builder{crd: &CustomResource{object: &apiext.CustomResourceDefinition{}}}
}

// WithName is used to add name field to the CRD
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD name"))
		return b
	}
	b.crd.object.Name = name
	return b
}

// WithGroup is used to add group field to the CRD
func (b *Builder) WithGroup(groupName string) *Builder {
	if len(groupName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD group"))
		return b
	}
	b.crd.object.Spec.Group = groupName
	return b
}

// WithVersion is used to add version field to the CRD
func (b *Builder) WithVersion(version string) *Builder {
	if len(version) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD version"))
		return b
	}
	b.crd.object.Spec.Version = version
	return b
}

// WithKind is used to add kind field to the CRD
func (b *Builder) WithKind(kind string) *Builder {
	if len(kind) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD kind"))
		return b
	}
	b.crd.object.Spec.Names.Kind = kind
	return b
}

// WithListKind is used to add listkind field to the CRD
func (b *Builder) WithListKind(listKind string) *Builder {
	if len(listKind) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD list kind"))
		return b
	}
	b.crd.object.Spec.Names.ListKind = listKind
	return b
}

// WithPlural is used to add plural field to the CRD
func (b *Builder) WithPlural(plural string) *Builder {
	if len(plural) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD plural name"))
		return b
	}
	b.crd.object.Spec.Names.Plural = plural
	return b
}

// WithShortNames is used to add shortnames field to the CRD
func (b *Builder) WithShortNames(shortNames []string) *Builder {
	if len(shortNames) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD shortnames"))
		return b
	}
	b.crd.object.Spec.Names.ShortNames = shortNames
	return b
}

// WithScope is used to add scope field to the CRD
func (b *Builder) WithScope(scope apiext.ResourceScope) *Builder {
	b.crd.object.Spec.Scope = scope
	return b
}

// WithPrinterColumns is used to add printercolumns field to the CRD
func (b *Builder) WithPrinterColumns(columnName, columnType, jsonPath string) *Builder {
	if len(columnName) == 0 {
		b.errs = append(b.errs,
			errors.New("missing column name in additional printer columns"))
		return b
	}
	if len(columnType) == 0 {
		b.errs = append(b.errs,
			errors.New("missing column type in additional printer columns"))
		return b
	}
	if len(jsonPath) == 0 {
		b.errs = append(b.errs,
			errors.New("missing json path in additional printer columns"))
		return b
	}
	printerColumn := apiext.CustomResourceColumnDefinition{
		Name:     columnName,
		Type:     columnType,
		JSONPath: jsonPath,
	}
	b.crd.object.Spec.AdditionalPrinterColumns = append(b.crd.object.Spec.AdditionalPrinterColumns, printerColumn)
	return b
}

// WithPriorityPrinterColumns is used to add printercolumns field to the CRD with priority field
func (b *Builder) WithPriorityPrinterColumns(columnName, columnType, jsonPath string, fsType string, priority int32) *Builder {
	if len(columnName) == 0 {
		b.errs = append(b.errs,
			errors.New("missing column name in additional printer columns"))
		return b
	}
	if len(columnType) == 0 {
		b.errs = append(b.errs,
			errors.New("missing column type in additional printer columns"))
		return b
	}
	if len(jsonPath) == 0 {
		b.errs = append(b.errs,
			errors.New("missing json path in additional printer columns"))
		return b
	}

	printerColumn := apiext.CustomResourceColumnDefinition{
		Name:     columnName,
		Type:     columnType,
		JSONPath: jsonPath,
		Priority: priority,
	}
	b.crd.object.Spec.AdditionalPrinterColumns = append(b.crd.object.Spec.AdditionalPrinterColumns, printerColumn)
	return b
}

// Build returns the CustomResourceDefinition from the builder
func (b *Builder) Build() (*apiext.CustomResourceDefinition, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("%+v", b.errs)
	}
	return b.crd.GetAPIObject(), nil
}
