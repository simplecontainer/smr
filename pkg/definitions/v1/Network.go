package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type NetworkDefinition struct {
	Meta NetworkMeta `json:"meta" validate:"required"`
	Spec NetworkSpec `json:"spec" validate:"required"`
}

type NetworkMeta struct {
	Group string         `json:"group" validate:"required"`
	Name  string         `json:"name" validate:"required"`
	Owner commonv1.Owner `json:"owner"`
}

type NetworkSpec struct {
	Driver          string
	IPV4AddressPool string
}

func (network *NetworkDefinition) SetOwner(kind string, group string, name string) {
	network.Meta.Owner.Kind = kind
	network.Meta.Owner.Group = group
	network.Meta.Owner.Name = name
}

func (network *NetworkDefinition) GetOwner() commonv1.Owner {
	return network.Meta.Owner
}

func (network *NetworkDefinition) GetKind() string {
	return static.KIND_NETWORK
}

func (network *NetworkDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (network *NetworkDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, network)
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
