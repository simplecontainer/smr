package definitions

import (
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
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
	case static.KIND_CUSTOM:
		def = &v1.CustomDefinition{}
	case static.KIND_NETWORK:
		def = &v1.NetworkDefinition{}
	case static.KIND_SECRET:
		def = &v1.SecretDefinition{}
	default:
		def = nil
	}

	if def != nil {
		def.SetRuntime(&commonv1.Runtime{
			Owner: commonv1.Owner{},
			Node:  0,
		})
	}

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
		return obj, obj.AddLocal(format, bytes)
	} else {
		return obj, nil
	}
}
func (definition *Definition) Delete(format contracts.Format, obj contracts.ObjectInterface, kind string) (contracts.IDefinition, error) {
	err := obj.Find(format)

	fmt.Println("trying delete definition")

	if err != nil {
		return nil, err
	}

	if obj.Exists() {
		existing := NewImplementation(kind)
		err = existing.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return existing, err
		}

		if !existing.GetRuntime().GetOwner().IsEqual(definition.GetRuntime().GetOwner()) {
			return existing, errors.New("object has owner - direct modification not allowed")
		}

		if err != nil {
			return nil, err
		}

		_, err = obj.RemoveLocal(format)
		return existing, err
	} else {
		return nil, errors.New("object doesnt exist")
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

func (definition *Definition) GetMeta() commonv1.Meta {
	return definition.Definition.GetMeta()
}

func (definition *Definition) GetState() *commonv1.State {
	return definition.Definition.GetState()
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
			Node:  0,
		})
	}

	if definition.GetState() == nil {
		definition.Definition.SetState(&commonv1.State{
			Options: make([]*commonv1.Opts, 0),
		})
	}

	return err
}

func (definition *Definition) ToJson() ([]byte, error) {
	return definition.Definition.ToJson()
}

func (definition *Definition) ToJsonForUser() ([]byte, error) {
	bytes, err := definition.Definition.ToJson()

	if err != nil {
		return nil, err
	}

	patchJSON := []byte(`[
		{"op": "remove", "path": "/meta/runtime"}
	]`)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		panic(err)
	}

	modified, err := patch.Apply(bytes)

	if err != nil {
		return nil, err
	}

	return modified, nil
}

func (definition *Definition) ToJsonString() (string, error) {
	return definition.Definition.ToJsonString()
}

func (definition *Definition) Validate() (bool, error) {
	return definition.Definition.Validate()
}
