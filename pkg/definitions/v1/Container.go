package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type ContainerDefinition struct {
	Meta ContainerMeta `json:"meta"  validate:"required"`
	Spec ContainerSpec `json:"spec"  validate:"required"`
}

type ContainerMeta struct {
	Name   string            `validate:"required" json:"name"`
	Group  string            `validate:"required" json:"group"`
	Labels map[string]string `json:"labels"`
	Owner  commonv1.Owner    `json:"-"`
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

func (container *ContainerDefinition) SetOwner(kind string, group string, name string) {
	container.Meta.Owner.Kind = kind
	container.Meta.Owner.Group = group
	container.Meta.Owner.Name = name
}

func (container *ContainerDefinition) GetOwner() commonv1.Owner {
	return container.Meta.Owner
}

func (container *ContainerDefinition) GetKind() string {
	return static.KIND_CONTAINER
}

func (container *ContainerDefinition) ResolveReferences(obj contracts.ObjectInterface) ([]contracts.IDefinition, error) {
	return nil, nil
}

func (container *ContainerDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, container)
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
