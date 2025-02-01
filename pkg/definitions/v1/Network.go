package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type NetworkDefinition struct {
	Prefix string        `json:"prefix" validate:"required"`
	Meta   commonv1.Meta `json:"meta" validate:"required"`
	Spec   NetworkSpec   `json:"spec" validate:"required"`
}

type NetworkSpec struct {
	Driver          string
	IPV4AddressPool string
}

func (network *NetworkDefinition) GetPrefix() string {
	return network.Prefix
}

func (network *NetworkDefinition) SetRuntime(runtime *commonv1.Runtime) {
	network.Meta.Runtime = runtime
}

func (network *NetworkDefinition) GetRuntime() *commonv1.Runtime {
	return network.Meta.Runtime
}

func (network *NetworkDefinition) GetMeta() commonv1.Meta {
	return network.Meta
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

func (network *NetworkDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(network)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "network"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
}

func (network *NetworkDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(network)
	return string(bytes), err
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
