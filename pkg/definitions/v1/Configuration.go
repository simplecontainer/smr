package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type ConfigurationDefinition struct {
	Meta ConfigurationMeta `json:"meta" validate:"required"`
	Spec ConfigurationSpec `json:"spec" validate:"required"`
}

type ConfigurationMeta struct {
	Group string         `json:"group" validate:"required"`
	Name  string         `json:"name" validate:"required"`
	Owner commonv1.Owner `json:"owner"`
}

type ConfigurationSpec struct {
	Data map[string]string `json:"data"`
}

func (configuration *ConfigurationDefinition) SetOwner(kind string, group string, name string) {
	configuration.Meta.Owner.Kind = kind
	configuration.Meta.Owner.Group = group
	configuration.Meta.Owner.Name = name
}

func (configuration *ConfigurationDefinition) GetOwner() commonv1.Owner {
	return configuration.Meta.Owner
}

func (configuration *ConfigurationDefinition) GetKind() string {
	return static.KIND_CONFIGURATION
}

func (configuration *ConfigurationDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (configuration *ConfigurationDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, configuration)
}

func (configuration *ConfigurationDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(configuration)
	return bytes, err
}

func (configuration *ConfigurationDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(configuration)
	return string(bytes), err
}

func (configuration *ConfigurationDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(configuration)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "configuration"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
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
