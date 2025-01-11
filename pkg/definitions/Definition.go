package definitions

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(kind string) *Definition {
	return &Definition{
		Definition: NewImplementation(kind),
	}
}

func NewImplementation(kind string) contracts.IDefinition {
	switch kind {
	case static.KIND_GITOPS:
		return &v1.GitopsDefinition{}
	case static.KIND_CONTAINER:
		return &v1.ContainerDefinition{}
	case static.KIND_CONTAINERS:
		return &v1.ContainersDefinition{}
	case static.KIND_CONFIGURATION:
		return &v1.ConfigurationDefinition{}
	case static.KIND_RESOURCE:
		return &v1.ResourceDefinition{}
	case static.KIND_HTTPAUTH:
		return &v1.HttpAuthDefinition{}
	case static.KIND_CERTKEY:
		return &v1.CertKeyDefinition{}
	case static.KIND_NETWORK:
		return &v1.NetworkDefinition{}
	default:
		return nil
	}
}

func (definition *Definition) Apply(format contracts.Format, obj contracts.ObjectInterface, kind string) (contracts.ObjectInterface, error) {
	err := obj.Find(format)

	if err != nil {
		return obj, err
	}

	var bytes []byte
	bytes, err = definition.Definition.ToJson()

	if err != nil {
		return obj, err
	}

	if definition.Definition.GetOwner().IsEmpty() {
		if obj.Exists() {
			existing := NewImplementation(kind)
			err = existing.FromJson(obj.GetDefinitionByte())

			if err != nil {
				return obj, err
			}

			if !existing.GetOwner().IsEmpty() {
				return obj, errors.New("object has owner - direct modification not allowed")
			}
		}

		if obj.Diff(bytes) {
			return obj, obj.Add(format, bytes)
		} else {
			return obj, nil
		}
	} else {
		if obj.Diff(bytes) {
			return obj, obj.AddLocal(format, bytes)
		} else {
			return obj, nil
		}
	}
}
func (definition *Definition) Delete(format contracts.Format, obj contracts.ObjectInterface, kind string) (contracts.IDefinition, error) {
	err := obj.Find(format)

	if err != nil {
		return nil, err
	}

	if definition.Definition.GetOwner().IsEmpty() {
		if obj.Exists() {
			existing := NewImplementation(kind)
			err = existing.FromJson(obj.GetDefinitionByte())

			if err != nil {
				return nil, err
			}

			if !existing.GetOwner().IsEmpty() {
				return nil, errors.New("object has owner - direct modification not allowed")
			}

			_, err = obj.Remove(format)
			return existing, err
		} else {
			return nil, errors.New("object doesnt exist")
		}
	} else {
		if obj.Exists() {
			existing := NewImplementation(kind)
			err = existing.FromJson(obj.GetDefinitionByte())

			if err != nil {
				return nil, err
			}

			_, err = obj.RemoveLocal(format)
			return existing, err
		} else {
			return nil, errors.New("object doesnt exist")
		}
	}
}
func (definition *Definition) Changed(format contracts.Format, obj contracts.ObjectInterface) (bool, error) {
	err := obj.Find(format)

	if err != nil {
		return false, err
	}

	var bytes []byte
	bytes, err = definition.ToJson()

	if err != nil {
		return false, err
	}

	if obj.Exists() {
		return obj.Diff(bytes), nil
	} else {
		if len(bytes) == 0 {
			return true, nil
		} else {
			return true, errors.New("object doesnt exist")

		}
	}
}

func (definition *Definition) SetOwner(kind string, group string, name string) {
	definition.Definition.SetOwner(kind, group, name)
}

func (definition *Definition) GetOwner() commonv1.Owner {
	return definition.Definition.GetOwner()
}

func (definition *Definition) GetKind() string {
	return definition.Definition.GetKind()
}

func (definition *Definition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return definition.Definition.ResolveReferences(obj)
}

func (definition *Definition) FromJson(bytes []byte) error {
	return definition.Definition.FromJson(bytes)
}

func (definition *Definition) ToJson() ([]byte, error) {
	return definition.Definition.ToJson()
}

func (definition *Definition) ToJsonString() (string, error) {
	return definition.Definition.ToJsonString()
}

func (definition *Definition) ToJsonStringWithKind() (string, error) {
	return definition.Definition.ToJsonStringWithKind()
}

func (definition *Definition) Validate() (bool, error) {
	return definition.Definition.Validate()
}
