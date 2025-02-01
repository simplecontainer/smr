package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type GitopsDefinition struct {
	Prefix string        `json:"prefix" validate:"required"`
	Meta   commonv1.Meta `json:"meta" validate:"required"`
	Spec   GitopsSpec    `json:"spec" validate:"required"`
}

type GitopsSpec struct {
	RepoURL         string            `json:"repoURL"`
	Revision        string            `json:"revision"`
	DirectoryPath   string            `json:"directoryPath"`
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

func (gitops *GitopsDefinition) GetPrefix() string {
	return gitops.Prefix
}

func (gitops *GitopsDefinition) SetRuntime(runtime *commonv1.Runtime) {
	gitops.Meta.Runtime = runtime
}

func (gitops *GitopsDefinition) GetRuntime() *commonv1.Runtime {
	return gitops.Meta.Runtime
}

func (gitops *GitopsDefinition) GetMeta() commonv1.Meta {
	return gitops.Meta
}

func (gitops *GitopsDefinition) GetKind() string {
	return static.KIND_GITOPS
}

func (gitops *GitopsDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	references := make([]contracts.IDefinition, 0)

	if gitops.Spec.HttpAuthRef.Group != "" && gitops.Spec.HttpAuthRef.Name != "" {
		/*
			format := f.New("httpauth", gitops.Spec.HttpAuthRef.Group, gitops.Spec.HttpAuthRef.Name, "object")

			request, err := common.NewRequest(static.KIND_HTTPAUTH)

			if err != nil {
				return references, err
			}

			err = request.Resolve(obj, format)

			if err != nil {
				return references, err
			}
		*/
	}

	if gitops.Spec.CertKeyRef.Group != "" && gitops.Spec.CertKeyRef.Name != "" {
		/*
			format := f.New("certkey", gitops.Spec.CertKeyRef.Group, gitops.Spec.CertKeyRef.Name, "object")

			request, err := common.NewRequest(static.KIND_CERTKEY)

			if err != nil {
				return references, err
			}

			err = request.Resolve(obj, format)

			if err != nil {
				return references, err
			}
		*/
	}

	return references, nil
}

func (gitops *GitopsDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, gitops)
}

func (gitops *GitopsDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(gitops)
	return bytes, err
}

func (gitops *GitopsDefinition) ToJsonWithKind() ([]byte, error) {
	bytes, err := json.Marshal(gitops)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return nil, err
	}

	definition["kind"] = "gitops"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return nil, err
	}

	return marshalled, err
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
