package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type HttpAuthDefinition struct {
	Meta commonv1.Meta `json:"meta" validate:"required"`
	Spec HttpAuthSpec  `json:"spec" validate:"required"`
}

type HttpAuthSpec struct {
	Username string
	Password string
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

func (httpauth *HttpAuthDefinition) GetKind() string {
	return static.KIND_HTTPAUTH
}

func (httpauth *HttpAuthDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (httpauth *HttpAuthDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, httpauth)
}

func (httpauth *HttpAuthDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(httpauth)
	return bytes, err
}

func (httpauth *HttpAuthDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(httpauth)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "httpauth"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
}

func (httpauth *HttpAuthDefinition) ToJsonString() (string, error) {
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
