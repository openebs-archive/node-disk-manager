package crds

import (
	"errors"
	"fmt"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type Builder struct {
	crd  *CustomResource
	errs []error
}

func NewBuilder() *Builder {
	return &Builder{crd: &CustomResource{object: &apiext.CustomResourceDefinition{}}}
}

func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD name"))
		return b
	}
	b.crd.object.Name = name
	return b
}

func (b *Builder) WithGroup(groupName string) *Builder {
	if len(groupName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD group"))
		return b
	}
	b.crd.object.Spec.Group = groupName
	return b
}

func (b *Builder) WithVersion(version string) *Builder {
	if len(version) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD version"))
		return b
	}
	b.crd.object.Spec.Version = version
	return b
}

func (b *Builder) WithKind(kind string) *Builder {
	if len(kind) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD kind"))
		return b
	}
	b.crd.object.Spec.Names.Kind = kind
	return b
}

func (b *Builder) WithListKind(listKind string) *Builder {
	if len(listKind) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD list kind"))
		return b
	}
	b.crd.object.Spec.Names.ListKind = listKind
	return b
}

func (b *Builder) WithPlural(plural string) *Builder {
	if len(plural) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD plural name"))
		return b
	}
	b.crd.object.Spec.Names.Plural = plural
	return b
}

func (b *Builder) WithShortNames(shortNames []string) *Builder {
	if len(shortNames) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CRD. missing CRD shortnames"))
		return b
	}
	b.crd.object.Spec.Names.ShortNames = shortNames
	return b
}

func (b *Builder) WithScope(scope apiext.ResourceScope) *Builder {
	b.crd.object.Spec.Scope = scope
	return b
}

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

func (b *Builder) Build() (*apiext.CustomResourceDefinition, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("%+v", b.errs)
	}
	return b.crd.object, nil
}
