package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type HttpAuthDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   commonv1.Meta   `json:"meta" validate:"required"`
	Spec   HttpAuthSpec    `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

type HttpAuthSpec struct {
	Username string
	Password string
}

func (httpauth *HttpAuthDefinition) GetPrefix() string {
	return httpauth.Prefix
}

func (httpauth *HttpAuthDefinition) SetRuntime(runtime *commonv1.Runtime) {
	httpauth.Meta.Runtime = runtime
}

func (httpauth *HttpAuthDefinition) GetRuntime() *commonv1.Runtime {
	return httpauth.Meta.Runtime
}

func (httpauth *HttpAuthDefinition) GetMeta() commonv1.Meta {
	return httpauth.Meta
}

func (httpauth *HttpAuthDefinition) GetState() *commonv1.State {
	return httpauth.State
}

func (httpauth *HttpAuthDefinition) SetState(state *commonv1.State) {
	httpauth.State = state
}

func (httpauth *HttpAuthDefinition) GetKind() string {
	return static.KIND_HTTPAUTH
}

func (httpauth *HttpAuthDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (httpauth *HttpAuthDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, httpauth)
}

func (httpauth *HttpAuthDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(httpauth)
	return bytes, err
}

func (httpauth *HttpAuthDefinition) ToJSONString() (string, error) {
	bytes, err := json.Marshal(httpauth)
	return string(bytes), err
}

func (httpauth *HttpAuthDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(httpauth)
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
