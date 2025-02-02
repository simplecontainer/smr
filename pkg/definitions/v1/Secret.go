package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type SecretDefinition struct {
	Prefix string        `json:"prefix" validate:"required"`
	Meta   commonv1.Meta `json:"meta" validate:"required"`
	Spec   SecretSpec    `json:"spec" validate:"required"`
}

type SecretSpec struct {
	Data map[string]string `json:"data" validate:"required"`
}

func (secret *SecretDefinition) GetPrefix() string {
	return secret.Prefix
}

func (secret *SecretDefinition) SetRuntime(runtime *commonv1.Runtime) {
	secret.Meta.Runtime = runtime
}

func (secret *SecretDefinition) GetRuntime() *commonv1.Runtime {
	return secret.Meta.Runtime
}

func (secret *SecretDefinition) GetMeta() commonv1.Meta {
	return secret.Meta
}

func (secret *SecretDefinition) GetKind() string {
	return static.KIND_SECRET
}

func (secret *SecretDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (secret *SecretDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, secret)
}

func (secret *SecretDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(secret)
	return bytes, err
}

func (secret *SecretDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(secret)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "secret"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
}

func (secret *SecretDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(secret)
	return string(bytes), err
}

func (secret *SecretDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(secret)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
