package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"gopkg.in/yaml.v3"
)

type CommonDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta" validate:"required"`
	Spec   CommonSpec      `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

type CommonSpec struct {
	Data map[string]any
}

func NewCommon() *CommonDefinition {
	return &CommonDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  CommonSpec{},
		State: nil,
	}
}

func (common *CommonDefinition) SetRuntime(runtime *commonv1.Runtime) {
	common.Meta.Runtime = runtime
}

func (common *CommonDefinition) GetRuntime() *commonv1.Runtime {
	return common.Meta.Runtime
}

func (common *CommonDefinition) GetPrefix() string {
	return common.Prefix
}

func (common *CommonDefinition) GetMeta() *commonv1.Meta {
	return common.Meta
}

func (common *CommonDefinition) GetState() *commonv1.State {
	return common.State
}

func (common *CommonDefinition) SetState(state *commonv1.State) {
	common.State = state
}

func (common *CommonDefinition) GetKind() string {
	return common.Kind
}

func (common *CommonDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (common *CommonDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, common)
}

func (common *CommonDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(common)
	return bytes, err
}

func (common *CommonDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(common)
	return bytes, err
}

func (common *CommonDefinition) ToJSONString() (string, error) {
	bytes, err := json.Marshal(common)
	return string(bytes), err
}

func (common *CommonDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(common)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
