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
	Meta HttpAuthMeta `json:"meta" validate:"required"`
	Spec HttpAuthSpec `json:"spec" validate:"required"`
}

type HttpAuthMeta struct {
	Group string         `json:"group" validate:"required"`
	Name  string         `json:"name" validate:"required"`
	Owner commonv1.Owner `json:"-"`
}

type HttpAuthSpec struct {
	Username string
	Password string
}

func (httpauth *HttpAuthDefinition) SetOwner(kind string, group string, name string) {
	httpauth.Meta.Owner.Kind = kind
	httpauth.Meta.Owner.Group = group
	httpauth.Meta.Owner.Name = name
}

func (httpauth *HttpAuthDefinition) GetOwner() commonv1.Owner {
	return httpauth.Meta.Owner
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

func (httpauth *HttpAuthDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(httpauth)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "httpauth"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
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
