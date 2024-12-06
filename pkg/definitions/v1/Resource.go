package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type ResourceDefinition struct {
	Meta ResourceMeta `json:"meta" validate:"required"`
	Spec ResourceSpec `json:"spec" validate:"required"`
}

type ResourceMeta struct {
	Group string `json:"group" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type ResourceSpec struct {
	Data map[string]string `json:"data"`
}

func (resource *ResourceDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(resource)
	return bytes, err
}

func (resource *ResourceDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

func (resource *ResourceDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(resource)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "resource"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
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
