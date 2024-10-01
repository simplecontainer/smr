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

func (definition *ContainersDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (definition *ContainersDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(definition)
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
