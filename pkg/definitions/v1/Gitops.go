package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/static"
	"gopkg.in/yaml.v3"
)

type GitopsDefinition struct {
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   *commonv1.Meta  `json:"meta" validate:"required"`
	Spec   *GitopsSpec     `json:"spec" validate:"required"`
	State  *commonv1.State `json:"state"`
}

type GitopsSpec struct {
	RepoURL         string             `json:"repoURL"`
	Revision        string             `json:"revision"`
	DirectoryPath   string             `json:"directoryPath"`
	PoolingInterval string             `json:"poolingInterval"`
	AutomaticSync   bool               `json:"automaticSync"`
	API             string             `json:"API"`
	Context         string             `json:"context"`
	CertKeyRef      *GitopsCertKeyRef  `json:"certKeyRef"`
	HttpAuthRef     *GitopsHttpauthRef `json:"httpAuthRef"`
}

type GitopsCertKeyRef struct {
	Prefix string
	Group  string
	Name   string
}

type GitopsHttpauthRef struct {
	Prefix string
	Group  string
	Name   string
}

func NewGitops() *GitopsDefinition {
	return &GitopsDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  &GitopsSpec{},
		State: nil,
	}
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

func (gitops *GitopsDefinition) GetMeta() *commonv1.Meta {
	return gitops.Meta
}

func (gitops *GitopsDefinition) GetState() *commonv1.State {
	return gitops.State
}

func (gitops *GitopsDefinition) SetState(state *commonv1.State) {
	gitops.State = state
}

func (gitops *GitopsDefinition) GetKind() string {
	return static.KIND_GITOPS
}

func (gitops *GitopsDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	references := make([]idefinitions.IDefinition, 0)

	if gitops.Spec.HttpAuthRef != nil {
		format := f.New(gitops.Spec.HttpAuthRef.Prefix, "kind", static.KIND_HTTPAUTH, gitops.Spec.HttpAuthRef.Group, gitops.Spec.HttpAuthRef.Name)
		obj.Find(format)

		if !obj.Exists() {
			return references, errors.New("gitops reference httpauth not found")
		}

		httpauth := &HttpAuthDefinition{}

		err := json.Unmarshal(obj.GetDefinitionByte(), httpauth)

		if err != nil {
			return references, err
		}

		references = append(references, httpauth)
	}

	if gitops.Spec.CertKeyRef != nil {
		fmt.Println(gitops.Spec.CertKeyRef)

		format := f.New(gitops.Spec.CertKeyRef.Prefix, "kind", static.KIND_CERTKEY, gitops.Spec.CertKeyRef.Group, gitops.Spec.CertKeyRef.Name)

		obj.Find(format)

		if !obj.Exists() {
			return references, errors.New("gitops reference certkey not found")
		}

		certkey := &CertKeyDefinition{}

		err := json.Unmarshal(obj.GetDefinitionByte(), certkey)

		if err != nil {
			return references, err
		}

		references = append(references, certkey)
	}

	return references, nil
}

func (gitops *GitopsDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, gitops)
}

func (gitops *GitopsDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(gitops)
	return bytes, err
}

func (gitops *GitopsDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(gitops)
	return bytes, err
}

func (gitops *GitopsDefinition) ToJSONString() (string, error) {
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
