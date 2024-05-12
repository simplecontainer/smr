package v1

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/qdnqn/smr/pkg/network"
)

type Containers struct {
	Kind       string               `json:"kind"  validate:"required"`
	Containers map[string]Container `json:"container"  validate:"required"`
}

type Container struct {
	Meta Meta `json:"meta"  validate:"required"`
	Spec Spec `json:"spec"  validate:"required"`
}

type Meta struct {
	Enabled bool              `json:"enabled"`
	Name    string            `json:"name"  validate:"required"`
	Group   string            `json:"group"  validate:"required"`
	Labels  map[string]string `json:"labels"  validate:"required"`
}

type Spec struct {
	Container ContainerInternal `json:"container"`
}

type ContainerInternal struct {
	Image         string                 `json:"image" validate:"required"`
	Tag           string                 `json:"tag"  validate:"required"`
	Envs          []string               `json:"envs"`
	Entrypoint    []string               `json:"entrypoint"`
	Command       []string               `json:"command"`
	Dependencies  []DependsOn            `json:"dependencies"`
	Networks      []string               `json:"networks" validate:"required"`
	Ports         []network.PortMappings `json:"ports"`
	Volumes       []map[string]string    `json:"volumes"`
	Operators     []map[string]any       `json:"operators"`
	Configuration map[string]any         `json:"configuration"`
	Resources     []map[string]any       `json:"resources"`
	Replicas      int                    `json:"replicas"  validate:"required"`
	Capabilities  []string               `json:"capabilities"`
	Privileged    bool                   `json:"privileged"`
	NetworkMode   string                 `json:"network_mode"`
}

type DependsOn struct {
	Name     string         `json:"name"  validate:"required"`
	Operator string         `json:"operator"`
	Timeout  string         `json:"timeout"  validate:"required"`
	Body     map[string]any `json:"body"`
	Solved   bool
}

func (definition *Container) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (definition *Container) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(definition)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}
