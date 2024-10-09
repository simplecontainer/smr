package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type HttpAuthDefinition struct {
	Meta HttpAuthMeta `json:"meta" validate:"required"`
	Spec HttpAuthSpec `json:"spec" validate:"required"`
}

type HttpAuthMeta struct {
	Group string `json:"group" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type HttpAuthSpec struct {
	Username string
	Password string
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
