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
	Meta GitopsMeta `json:"meta" validate:"required"`
	Spec GitopsSpec `json:"spec" validate:"required"`
}

type GitopsMeta struct {
	Group string         `json:"group" validate:"required"`
	Name  string         `json:"name" validate:"required"`
	Owner commonv1.Owner `json:"-"`
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

func (gitops *GitopsDefinition) SetOwner(kind string, group string, name string) {
	gitops.Meta.Owner.Kind = kind
	gitops.Meta.Owner.Group = group
	gitops.Meta.Owner.Name = name
}

func (gitops *GitopsDefinition) GetOwner() commonv1.Owner {
	return gitops.Meta.Owner
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

func (gitops *GitopsDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(gitops)
	return string(bytes), err
}

func (gitops *GitopsDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(gitops)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "gitops"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
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
