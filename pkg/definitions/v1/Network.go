package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
)

type NetworkDefinition struct {
	Meta NetworkMeta `json:"meta" validate:"required"`
	Spec NetworkSpec `json:"spec" validate:"required"`
}

type NetworkMeta struct {
	Group string `json:"group" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type NetworkSpec struct {
	Driver          string
	IPV4AddressPool string
}

func (Network *NetworkDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(Network)
	return string(bytes), err
}

func (Network *NetworkDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(Network)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
