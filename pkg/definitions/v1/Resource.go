package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type ResourceDefinition struct {
	Meta ResourceMeta `json:"meta" validate:"required"`
	Spec ResourceSpec `json:"spec" validate:"required"`
}

type ResourceMeta struct {
	Group   string            `json:"group" validate:"required"`
	Name    string            `json:"name" validate:"required"`
	Runtime *commonv1.Runtime `json:"runtime"`
}

type ResourceSpec struct {
	Data map[string]string `json:"data"`
}

func (resource *ResourceDefinition) SetRuntime(runtime *commonv1.Runtime) {
	resource.Meta.Runtime = runtime
}

func (resource *ResourceDefinition) GetRuntime() *commonv1.Runtime {
	return resource.Meta.Runtime
}

func (resource *ResourceDefinition) GetKind() string {
	return static.KIND_RESOURCE
}

func (resource *ResourceDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (resource *ResourceDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, resource)
}

func (resource *ResourceDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(resource)
	return bytes, err
}

func (resource *ResourceDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(resource)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "resource"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
}

func (resource *ResourceDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

func (resource *ResourceDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(resource)
	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
