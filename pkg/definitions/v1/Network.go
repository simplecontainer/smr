package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
	"gopkg.in/yaml.v3"
)

type NetworkDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta" validate:"required"`
	Spec   NetworkSpec     `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

type NetworkSpec struct {
	Driver          string
	IPV4AddressPool string
}

func NewNetwork() *NetworkDefinition {
	return &NetworkDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  NetworkSpec{},
		State: nil,
	}
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

func (network *NetworkDefinition) GetMeta() *commonv1.Meta {
	return network.Meta
}

func (network *NetworkDefinition) GetState() *commonv1.State {
	return network.State
}

func (network *NetworkDefinition) SetState(state *commonv1.State) {
	network.State = state
}

func (network *NetworkDefinition) GetKind() string {
	return static.KIND_NETWORK
}

func (network *NetworkDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (network *NetworkDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, network)
}

func (network *NetworkDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(network)
	return bytes, err
}

func (network *NetworkDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(network)
	return bytes, err
}

func (network *NetworkDefinition) ToJSONString() (string, error) {
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
