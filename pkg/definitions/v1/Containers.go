package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type ContainersDefinition struct {
	Kind string                         `json:"kind"  validate:"required"`
	Meta ContainersMeta                 `json:"meta"  validate:"required"`
	Spec map[string]ContainerDefinition `json:"spec"  validate:"required"`
}

type ContainersMeta struct {
	Enabled bool              `json:"enabled"`
	Name    string            `validate:"required" json:"name"`
	Group   string            `validate:"required" json:"group"`
	Labels  map[string]string `json:"labels"`
}

func (containers *ContainersDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(containers)
	return bytes, err
}

func (containers *ContainersDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(containers)
	return string(bytes), err
}

func (containers *ContainersDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(containers)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "containers"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
}

func (containers *ContainersDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(containers)
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
