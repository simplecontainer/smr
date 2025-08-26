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

type CustomDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta" validate:"required"`
	Spec   map[string]any  `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

func NewCustom() *CustomDefinition {
	return &CustomDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  map[string]any{},
		State: nil,
	}
}

func (custom *CustomDefinition) GetPrefix() string {
	return custom.Prefix
}

func (custom *CustomDefinition) SetRuntime(runtime *commonv1.Runtime) {
	custom.Meta.Runtime = runtime
}

func (custom *CustomDefinition) GetRuntime() *commonv1.Runtime {
	return custom.Meta.Runtime
}

func (custom *CustomDefinition) GetMeta() *commonv1.Meta {
	return custom.Meta
}

func (custom *CustomDefinition) GetState() *commonv1.State {
	return custom.State
}

func (custom *CustomDefinition) SetState(state *commonv1.State) {
	custom.State = state
}

func (custom *CustomDefinition) GetKind() string {
	return static.KIND_CUSTOM
}

func (custom *CustomDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (custom *CustomDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, custom)
}

func (custom *CustomDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(custom)
	return bytes, err
}

func (custom *CustomDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(custom)
	return bytes, err
}

func (custom *CustomDefinition) ToJSONString() (string, error) {
	bytes, err := json.Marshal(custom)
	return string(bytes), err
}

func (custom *CustomDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(custom)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
