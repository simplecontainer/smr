package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
	"gopkg.in/yaml.v3"
)

type ResourceDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta" validate:"required"`
	Spec   ResourceSpec    `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

type ResourceSpec struct {
	Data map[string]string `json:"data"`
}

func NewResource() *ResourceDefinition {
	return &ResourceDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  ResourceSpec{},
		State: nil,
	}
}

func (resource *ResourceDefinition) GetPrefix() string {
	return resource.Prefix
}

func (resource *ResourceDefinition) SetRuntime(runtime *commonv1.Runtime) {
	resource.Meta.Runtime = runtime
}

func (resource *ResourceDefinition) GetRuntime() *commonv1.Runtime {
	return resource.Meta.Runtime
}

func (resource *ResourceDefinition) GetMeta() *commonv1.Meta {
	return resource.Meta
}

func (resource *ResourceDefinition) GetState() *commonv1.State {
	return resource.State
}

func (resource *ResourceDefinition) SetState(state *commonv1.State) {
	resource.State = state
}

func (resource *ResourceDefinition) GetKind() string {
	return static.KIND_RESOURCE
}

func (resource *ResourceDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (resource *ResourceDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, resource)
}

func (resource *ResourceDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(resource)
	return bytes, err
}

func (resource *ResourceDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(resource)
	return bytes, err
}

func (resource *ResourceDefinition) ToJSONString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

func (resource *ResourceDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(resource)
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
