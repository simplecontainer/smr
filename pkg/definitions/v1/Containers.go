package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type ContainersDefinition struct {
	Kind   string                         `json:"kind" validate:"required"`
	Prefix string                         `json:"prefix" validate:"required"`
	Meta   commonv1.Meta                  `json:"meta"  validate:"required"`
	Spec   map[string]ContainerDefinition `json:"spec"  validate:"required"`
	State  *commonv1.State                `json:"state"`
}

func (containers *ContainersDefinition) GetPrefix() string {
	return containers.Prefix
}

func (containers *ContainersDefinition) SetRuntime(runtime *commonv1.Runtime) {
	containers.Meta.Runtime = runtime
}

func (containers *ContainersDefinition) GetRuntime() *commonv1.Runtime {
	return containers.Meta.Runtime
}

func (containers *ContainersDefinition) GetMeta() commonv1.Meta {
	return containers.Meta
}

func (containers *ContainersDefinition) GetState() *commonv1.State {
	return containers.State
}

func (containers *ContainersDefinition) SetState(state *commonv1.State) {
	containers.State = state
}

func (containers *ContainersDefinition) GetKind() string {
	return static.KIND_CONTAINERS
}

func (containers *ContainersDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (containers *ContainersDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, containers)
}

func (containers *ContainersDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(containers)
	return bytes, err
}

func (containers *ContainersDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(containers)
	return string(bytes), err
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
