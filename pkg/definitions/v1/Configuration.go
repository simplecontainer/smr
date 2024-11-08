package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
)

type ConfigurationDefinition struct {
	Meta ConfigurationMeta `json:"meta" validate:"required"`
	Spec ConfigurationSpec `json:"spec" validate:"required"`
}

type ConfigurationMeta struct {
	Group string `json:"group" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type ConfigurationSpec struct {
	Data map[string]string `json:"data"`
}

func (configuration *ConfigurationDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(configuration)
	return bytes, err
}

func (configuration *ConfigurationDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(configuration)
	return string(bytes), err
}

func (configuration *ConfigurationDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(configuration)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
