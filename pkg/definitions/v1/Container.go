package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
)

type ContainerDefinition struct {
	Meta ContainerMeta `json:"meta"  validate:"required"`
	Spec ContainerSpec `json:"spec"  validate:"required"`
}

type ContainerMeta struct {
	Name   string            `validate:"required" json:"name"`
	Group  string            `validate:"required" json:"group"`
	Labels map[string]string `json:"labels"`
	Owner  ContainerOwner    `json:"-"`
}

type ContainerOwner struct {
	Kind  string
	Group string
	Name  string
}

func (owner ContainerOwner) IsEmpty() bool {
	return owner.Group != "" && owner.Name != ""
}

type ContainerSpec struct {
	Container ContainerInternal `validate:"required" json:"container" `
}

type ContainerInternal struct {
	Image         string               `validate:"required" json:"image"`
	Tag           string               `validate:"required" json:"tag"`
	Envs          []string             `json:"envs"`
	Entrypoint    []string             `json:"entrypoint"`
	Args          []string             `json:"args"`
	Dependencies  []ContainerDependsOn `json:"dependencies"`
	Readiness     []ContainerReadiness `json:"readiness"`
	Networks      []ContainerNetwork   `json:"networks"`
	Ports         []ContainerPort      `json:"ports"`
	Volumes       []ContainerVolume    `json:"volumes"`
	Configuration map[string]string    `json:"configuration"`
	Resources     []ContainerResource  `json:"resources"`
	Replicas      uint64               `validate:"required" json:"replicas"`
	Capabilities  []string             `json:"capabilities"`
	Privileged    bool                 `json:"privileged"`
	NetworkMode   string               `json:"network_mode"`
	Spread        ContainerSpread      `json:"spread"`
	Nodes         []string             `json:"nodes"`
	Dns           []string             `json:"dns"`
}

type ContainerDependsOn struct {
	Name    string `validate:"required" json:"name"`
	Group   string `validate:"required" json:"group"`
	Timeout string `validate:"required" json:"timeout"`
}

type ContainerReadiness struct {
	Name    string   `validate:"required" json:"name"`
	Type    string   `json:"type"`
	URL     string   `json:"url"`
	Command []string `json:"command"`
	Timeout string   `validate:"required" json:"timeout"`
}

type ContainerSpread struct {
	Spread string   `json:"spread"`
	Agents []uint64 `json:"agents"`
}

type ContainerNetwork struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

type ContainerPort struct {
	Container string `json:"container"`
	Host      string `json:"host"`
}

type ContainerVolume struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	HostPath   string `json:"hostPath"`
	MountPoint string `json:"mountPoint"`
}

type ContainerResource struct {
	Name       string
	Group      string
	Key        string
	MountPoint string
}

func (container *ContainerDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(container)
	return bytes, err
}

func (container *ContainerDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(container)
	return string(bytes), err
}

func (container *ContainerDefinition) ToJsonStringWithKind() (string, error) {
	bytes, err := json.Marshal(container)

	var definition map[string]interface{}
	err = json.Unmarshal(bytes, &definition)

	if err != nil {
		return "", err
	}

	definition["kind"] = "container"

	var marshalled []byte
	marshalled, err = json.Marshal(definition)

	if err != nil {
		return "", err
	}

	return string(marshalled), err
}

func (container *ContainerDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(container)
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
