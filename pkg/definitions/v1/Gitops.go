package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type GitopsDefinition struct {
	Meta GitopsMeta `json:"meta"`
	Spec GitopsSpec `json:"spec"`
}

type GitopsMeta struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

type GitopsSpec struct {
	RepoURL         string            `json:"repoURL"`
	Revision        string            `json:"revision"`
	DirectoryPath   string            `json:"directory"`
	PoolingInterval string            `json:"poolingInterval"`
	AutomaticSync   bool              `json:"automaticSync"`
	API             string            `json:"API"`
	Context         string            `json:"context"`
	CertKeyRef      GitopsCertKeyRef  `json:"certKeyRef"`
	HttpAuthRef     GitopsHttpauthRef `json:"httpAuthRef"`
}

type GitopsCertKeyRef struct {
	Group string
	Name  string
}

type GitopsHttpauthRef struct {
	Group string
	Name  string
}

func (gitops *GitopsDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(gitops)
	return string(bytes), err
}

func (gitops *GitopsDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(gitops)
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
