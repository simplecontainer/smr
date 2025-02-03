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
	Kind   string          `json:"kind" validate:"required"`
	Prefix string          `json:"prefix" validate:"required"`
	Meta   commonv1.Meta   `json:"meta"  validate:"required"`
	Spec   ContainerSpec   `json:"spec"  validate:"required"`
	State  *commonv1.State `json:"state"`
}

type ContainerSpec struct {
	Container ContainerInternal `validate:"required" json:"container" `
}

type ContainerInternal struct {
	Image         string               `validate:"required" json:"image"`
	Tag           string               `validate:"required" json:"tag"`
	Envs          []string             `json:"envs,omitempty"`
	Entrypoint    []string             `json:"entrypoint,omitempty"`
	Args          []string             `json:"args,omitempty"`
	Dependencies  []ContainerDependsOn `json:"dependencies,omitempty"`
	Readiness     []ContainerReadiness `json:"readiness,omitempty"`
	Networks      []ContainerNetwork   `json:"networks,omitempty"`
	Ports         []ContainerPort      `json:"ports,omitempty"`
	Volumes       []ContainerVolume    `json:"volumes,omitempty"`
	Configuration map[string]string    `json:"configuration,omitempty"`
	Resources     []ContainerResource  `json:"resources,omitempty"`
	Replicas      uint64               `validate:"required" json:"replicas"`
	Capabilities  []string             `json:"capabilities,omitempty"`
	Privileged    bool                 `json:"privileged,omitempty"`
	NetworkMode   string               `json:"network_mode,omitempty"`
	Spread        ContainerSpread      `json:"spread,omitempty"`
	Nodes         []string             `json:"nodes,omitempty"`
	Dns           []string             `json:"dns,omitempty"`
}

type ContainerDependsOn struct {
	Prefix  string `json:"prefix" default:"simplecontainer.io/v1"`
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
	Agents []uint64 `json:"agents,omitempty"`
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

func (container *ContainerDefinition) GetPrefix() string {
	return container.Prefix
}

func (container *ContainerDefinition) SetRuntime(runtime *commonv1.Runtime) {
	container.Meta.Runtime = runtime
}

func (container *ContainerDefinition) GetRuntime() *commonv1.Runtime {
	return container.Meta.Runtime
}

func (container *ContainerDefinition) GetMeta() commonv1.Meta {
	return container.Meta
}

func (container *ContainerDefinition) GetState() *commonv1.State {
	return container.State
}

func (container *ContainerDefinition) SetState(state *commonv1.State) {
	container.State = state
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
