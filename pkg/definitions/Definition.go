package definitions

import (
	"encoding/json"
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/static"
	"sync"
	"time"
)

func New(kind string) *Definition {
	return &Definition{
		Definition: NewImplementation(kind),
		Kind:       kind,
	}
}

func NewImplementation(kind string) idefinitions.IDefinition {
	var def idefinitions.IDefinition

	switch kind {
	case static.KIND_GITOPS:
		def = &v1.GitopsDefinition{}
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

		def.SetState(commonv1.NewState())
	}

	return def
}

func (definition *Definition) UnmarshalJSON(data []byte) error {
	var raw struct {
		Definition json.RawMessage `json:"definition"`
		Kind       string
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw.Kind {
	case static.KIND_GITOPS:
		tmp := &v1.GitopsDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_CONTAINERS:
		tmp := &v1.ContainersDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_CONFIGURATION:
		tmp := &v1.ConfigurationDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_RESOURCE:
		tmp := &v1.ResourceDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_HTTPAUTH:
		tmp := &v1.HttpAuthDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_CERTKEY:
		tmp := &v1.CertKeyDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_CUSTOM:
		tmp := &v1.CustomDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_NETWORK:
		tmp := &v1.NetworkDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	case static.KIND_SECRET:
		tmp := &v1.SecretDefinition{}

		err := json.Unmarshal(raw.Definition, tmp)
		if err != nil {
			return err
		}

		definition.Definition = tmp
	default:
		definition.Definition = nil
	}

	return nil
}

func (definition *Definition) Apply(format iformat.Format, obj iobjects.ObjectInterface) (iobjects.ObjectInterface, error) {
	err := obj.Find(format)

	if err != nil {
		return obj, err
	}

	var bytes []byte
	bytes, err = definition.Definition.ToJSON()

	if err != nil {
		return obj, err
	}

	if obj.Exists() {
		existing := NewImplementation(definition.GetKind())
		err = existing.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return obj, err
		}

		if !existing.GetRuntime().GetOwner().IsEqual(definition.GetRuntime().GetOwner()) {
			return obj, errors.New(fmt.Sprintf("object has owner - direct modification not allowed (%v)", definition.GetMeta()))
		}
	}

	if obj.Diff(bytes) {
		return obj, obj.AddLocal(format, bytes)
	} else {
		return obj, nil
	}
}
func (definition *Definition) State(format iformat.Format, obj iobjects.ObjectInterface) (iobjects.ObjectInterface, error) {
	err := obj.Find(format)

	if err != nil {
		return obj, err
	}

	if obj.Exists() {
		var bytes []byte
		bytes, err = definition.Definition.ToJSON()

		if err != nil {
			return obj, err
		}

		existing := NewImplementation(definition.GetKind())
		err = existing.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return obj, err
		}

		if !existing.GetRuntime().GetOwner().IsEqual(definition.GetRuntime().GetOwner()) {
			existing.SetState(definition.GetState())
		}

		bytes, err = existing.ToJSON()

		if err != nil {
			return obj, err
		}

		return obj, obj.AddLocal(format, bytes)
	}

	return obj, nil
}
func (definition *Definition) Delete(format iformat.Format, obj iobjects.ObjectInterface) (idefinitions.IDefinition, error) {
	err := obj.Find(format)

	if err != nil {
		return nil, err
	}

	if obj.Exists() {
		existing := NewImplementation(definition.GetKind())
		err = existing.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return existing, err
		}

		if !existing.GetRuntime().GetOwner().IsEqual(definition.GetRuntime().GetOwner()) {
			return existing, errors.New(fmt.Sprintf("object has owner - direct modification not allowed (%v)", definition.GetMeta()))
		}

		_, err = obj.RemoveLocal(format)
		return existing, err
	} else {
		return nil, errors.New("object doesnt exist")
	}
}
func (definition *Definition) Changed(format iformat.Format, obj iobjects.ObjectInterface) (bool, error) {
	err := obj.Find(format)

	if err != nil {
		return false, err
	}

	var bytes []byte
	bytes, err = definition.ToJSON()

	if err != nil {
		return false, err
	}

	if obj.Exists() {
		return obj.Diff(bytes), nil
	} else {
		if len(bytes) == 0 {
			return true, nil
		} else {
			return true, errors.New(static.RESPONSE_NOT_FOUND)
		}
	}
}

func (definition *Definition) SetRuntime(runtime *commonv1.Runtime) {
	definition.Definition.SetRuntime(runtime)
}

func (definition *Definition) GetRuntime() *commonv1.Runtime {
	return definition.Definition.GetRuntime()
}

func (definition *Definition) GetPrefix() string {
	return definition.Definition.GetPrefix()
}

func (definition *Definition) GetMeta() commonv1.Meta {
	return definition.Definition.GetMeta()
}

func (definition *Definition) GetState() *commonv1.State {
	return definition.Definition.GetState()
}

func (definition *Definition) SetState(state *commonv1.State) {
	definition.Definition.SetState(state)
}

func (definition *Definition) GetKind() string {
	return definition.Definition.GetKind()
}

func (definition *Definition) IsOf(compare idefinitions.IDefinition) bool {
	if definition.GetKind() == compare.GetKind() &&
		definition.GetMeta().Group == compare.GetMeta().Group &&
		definition.GetMeta().Name == compare.GetMeta().Name {
		return true
	} else {
		return false
	}
}

// Patch is reversed -> compare is patched with definition
func (definition *Definition) Patch(current idefinitions.IDefinition) error {
	var newDefinition []byte
	var currentDefinition []byte
	var patch []byte
	var modified []byte
	var err error

	newDefinition, err = definition.Definition.ToJSON()

	if err != nil {
		return err
	}

	currentDefinition, err = current.ToJSON()

	if err != nil {
		return err
	}

	patch, err = jsonpatch.CreateMergePatch(currentDefinition, newDefinition)

	if err != nil {
		return err
	}

	modified, err = jsonpatch.MergePatch(currentDefinition, patch)

	if err != nil {
		return err
	}

	return definition.FromJson(modified)
}

func (definition *Definition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
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
			Lock:    &sync.RWMutex{},
			Options: make([]*commonv1.Opts, 0),
			Gitops: commonv1.Gitops{
				Synced:   false,
				Drifted:  false,
				Missing:  false,
				NotOwner: false,
				Error:    false,
				State:    "",
				Messages: nil,
				Commit:   plumbing.Hash{},
				Changes:  nil,
				LastSync: time.Time{},
			},
		})
	}

	return err
}

func (definition *Definition) ToJSON() ([]byte, error) {
	return definition.Definition.ToJSON()
}

func (definition *Definition) ToJSONForUser() ([]byte, error) {
	bytes, err := definition.Definition.ToJSON()

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

func (definition *Definition) ToJSONString() (string, error) {
	return definition.Definition.ToJSONString()
}

func (definition *Definition) Validate() (bool, error) {
	return definition.Definition.Validate()
}
