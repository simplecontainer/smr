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

func (network *NetworkDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(network)
	return bytes, err
}

func (network *NetworkDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(network)
	return string(bytes), err
}

func (network *NetworkDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(network)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "network"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
}

func (network *NetworkDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(network)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
