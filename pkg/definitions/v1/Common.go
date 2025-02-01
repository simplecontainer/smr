package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type CommonDefinition struct {
	Prefix string        `json:"prefix" validate:"required"`
	Meta   commonv1.Meta `json:"meta" validate:"required"`
	Spec   CommonSpec    `json:"spec" validate:"required"`
}

type CommonSpec struct {
	Data map[string]any
}

func (common *CommonDefinition) SetRuntime(runtime *commonv1.Runtime) {
	common.Meta.Runtime = runtime
}

func (common *CommonDefinition) GetRuntime() *commonv1.Runtime {
	return common.Meta.Runtime
}

func (common *CommonDefinition) GetPrefix() string {
	return common.Prefix
}

func (common *CommonDefinition) GetMeta() commonv1.Meta {
	return common.Meta
}

func (common *CommonDefinition) GetKind() string {
	return static.KIND_CERTKEY
}

func (common *CommonDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (common *CommonDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, common)
}

func (common *CommonDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(common)
	return bytes, err
}

func (common *CommonDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(common)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "common"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
}

func (common *CommonDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(common)
	return string(bytes), err
}

func (common *CommonDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(common)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
