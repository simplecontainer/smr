package definitions

import (
	"errors"
	"fmt"
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
	var def contracts.IDefinition

	switch kind {
	case static.KIND_GITOPS:
		def = &v1.GitopsDefinition{}
	case static.KIND_CONTAINER:
		def = &v1.ContainerDefinition{}
	case static.KIND_CONTAINERS:
		def = &v1.ContainersDefinition{}
	case static.KIND_CONFIGURATION:
		def = &v1.ConfigurationDefinition{}
	case static.KIND_RESOURCE:
		def = &v1.ResourceDefinition{}
	case static.KIND_HTTPAUTH:
		def = &v1.HttpAuthDefinition{}
	case static.KIND_CERTKEY:
		def = &v1.CertKeyDefinition{}
	case static.KIND_NETWORK:
		def = &v1.NetworkDefinition{}
	default:
		def = nil
	}

	def.SetRuntime(&commonv1.Runtime{
		Owner: commonv1.Owner{},
		Node:  "",
	})

	return def
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

	if obj.Exists() {
		existing := NewImplementation(kind)
		err = existing.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return obj, err
		}

		if !existing.GetRuntime().GetOwner().IsEqual(definition.GetRuntime().GetOwner()) {
			return obj, errors.New("object has owner - direct modification not allowed")
		}
	}

	if obj.Diff(bytes) {
		fmt.Println("adding local")
		fmt.Println(format)

		return obj, obj.AddLocal(format, bytes)
	} else {
		return obj, nil
	}
}

func (definition *Definition) Delete(format contracts.Format, obj contracts.ObjectInterface, kind string) (contracts.IDefinition, error) {
	err := obj.Find(format)

	if err != nil {
		return nil, err
	}

	if definition.Definition.GetRuntime().GetOwner().IsEmpty() {
		if obj.Exists() {
			existing := NewImplementation(kind)
			err = existing.FromJson(obj.GetDefinitionByte())

			if err != nil {
				return nil, err
			}

			if !existing.GetRuntime().GetOwner().IsEmpty() {
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

func (definition *Definition) SetRuntime(runtime *commonv1.Runtime) {
	definition.Definition.SetRuntime(runtime)
}

func (definition *Definition) GetRuntime() *commonv1.Runtime {
	return definition.Definition.GetRuntime()
}

func (definition *Definition) GetKind() string {
	return definition.Definition.GetKind()
}

func (definition *Definition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return definition.Definition.ResolveReferences(obj)
}

func (definition *Definition) FromJson(bytes []byte) error {
	err := definition.Definition.FromJson(bytes)

	// Protect if json unmarshal nilify runtime
	if definition.GetRuntime() == nil {
		definition.SetRuntime(&commonv1.Runtime{
			Owner: commonv1.Owner{},
			Node:  "",
		})
	}

	return err
}

func (definition *Definition) ToJson() ([]byte, error) {
	return definition.Definition.ToJson()
}

func (definition *Definition) ToJsonWithKind() ([]byte, error) {
	return definition.Definition.ToJsonWithKind()
}

func (definition *Definition) ToJsonString() (string, error) {
	return definition.Definition.ToJsonString()
}

func (definition *Definition) Validate() (bool, error) {
	return definition.Definition.Validate()
}
