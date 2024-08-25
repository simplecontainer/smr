package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type Containers struct {
	Kind string               `json:"kind"  validate:"required"`
	Meta Meta                 `json:"meta"  validate:"required"`
	Spec map[string]Container `json:"spec"  validate:"required"`
}

type Container struct {
	Meta Meta `json:"meta"  validate:"required"`
	Spec Spec `json:"spec"  validate:"required"`
}

type Meta struct {
	Enabled bool              `json:"enabled"`
	Name    string            `validate:"required" json:"name"`
	Group   string            `validate:"required" json:"group"`
	Labels  map[string]string `json:"labels"`
}

type Spec struct {
	Container ContainerInternal `validate:"required" json:"container" `
}

type ContainerInternal struct {
	Image         string              `validate:"required" json:"image"`
	Tag           string              `validate:"required" json:"tag"`
	Envs          []string            `json:"envs"`
	Entrypoint    []string            `json:"entrypoint"`
	Command       []string            `json:"command"`
	Dependencies  []DependsOn         `json:"dependencies"`
	Readiness     []Readiness         `json:"readiness"`
	Networks      []string            `validate:"required" json:"networks"`
	Ports         []map[string]string `json:"ports"`
	Volumes       []map[string]string `json:"volumes"`
	Operators     []map[string]any    `json:"operators"`
	Configuration map[string]string   `json:"configuration"`
	Resources     []map[string]string `json:"resources"`
	Replicas      int                 `validate:"required" json:"replicas"`
	Capabilities  []string            `json:"capabilities"`
	Privileged    bool                `json:"privileged"`
	NetworkMode   string              `json:"network_mode"`
}

type DependsOn struct {
	Name    string `validate:"required" json:"name"`
	Group   string `validate:"required" json:"group"`
	Timeout string `validate:"required" json:"timeout"`
}

type Readiness struct {
	Name     string            `validate:"required" json:"name"`
	Operator string            `json:"operator"`
	Timeout  string            `validate:"required" json:"timeout"`
	Body     map[string]string `json:"body"`
}

func (definition *Containers) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (definition *Container) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (definition *Container) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(definition)
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

func (definition *Containers) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(definition)
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
