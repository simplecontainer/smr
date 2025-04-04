package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
)

type ContainersDefinition struct {
	Kind          string             `json:"kind" validate:"required"`
	Prefix        string             `json:"prefix" validate:"required"`
	Meta          commonv1.Meta      `json:"meta"  validate:"required"`
	Spec          ContainersInternal `json:"spec"  validate:"required"`
	InitContainer ContainersInternal `json:"initContainer,omitempty" validate:"omitempty"`
	State         *commonv1.State    `json:"state"`
}

type ContainersInternal struct {
	Image          string                     `validate:"required" json:"image"`
	Tag            string                     `validate:"required" json:"tag"`
	RepositoryAuth string                     `json:"repositoryAuth"`
	Envs           []string                   `json:"envs,omitempty"`
	Entrypoint     []string                   `json:"entrypoint,omitempty"`
	Args           []string                   `json:"args,omitempty"`
	Dependencies   []ContainersDependsOn      `json:"dependencies,omitempty"`
	Readiness      []ContainersReadiness      `json:"readiness,omitempty"`
	Networks       []ContainersNetwork        `json:"networks,omitempty"`
	Ports          []ContainersPort           `json:"ports,omitempty"`
	Volumes        []ContainersVolume         `json:"volumes,omitempty"`
	Configuration  map[string]string          `json:"configuration,omitempty"`
	Resources      []ContainersResource       `json:"resources,omitempty"`
	Configurations []ContainersConfigurations `json:"configurations,omitempty"`
	Replicas       uint64                     `validate:"required" json:"replicas"`
	Capabilities   []string                   `json:"capabilities,omitempty"`
	Privileged     bool                       `json:"privileged,omitempty"`
	NetworkMode    string                     `json:"network_mode,omitempty"`
	Spread         ContainersSpread           `json:"spread,omitempty"`
	Nodes          []string                   `json:"nodes,omitempty"`
	Dns            []string                   `json:"dns,omitempty"`
}

type ContainersDependsOn struct {
	Prefix  string `json:"prefix" default:"simplecontainers.io/v1"`
	Name    string `validate:"required" json:"name"`
	Group   string `validate:"required" json:"group"`
	Timeout string `validate:"required" json:"timeout"`
}

type ContainersReadiness struct {
	Name    string            `validate:"required" json:"name"`
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Body    map[string]string `json:"body"`
	Method  string            `json:"method"`
	Command []string          `json:"command"`
	Timeout string            `validate:"required" json:"timeout"`
}

type ContainersSpread struct {
	Spread string   `json:"spread"`
	Agents []uint64 `json:"agents,omitempty"`
}

type ContainersNetwork struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

type ContainersPort struct {
	Container string `json:"container"`
	Host      string `json:"host"`
}

type ContainersVolume struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	HostPath   string `json:"hostPath"`
	MountPoint string `json:"mountPoint"`
	SubPath    string `json:"subPath" default:"/"`
}

type ContainersResource struct {
	Name       string
	Group      string
	Key        string
	MountPoint string
}

type ContainersConfigurations struct {
	Name  string
	Group string
}

func (containers *ContainersDefinition) GetPrefix() string {
	return containers.Prefix
}

func (containers *ContainersDefinition) SetRuntime(runtime *commonv1.Runtime) {
	containers.Meta.Runtime = runtime
}

func (containers *ContainersDefinition) GetRuntime() *commonv1.Runtime {
	return containers.Meta.Runtime
}

func (containers *ContainersDefinition) GetMeta() commonv1.Meta {
	return containers.Meta
}

func (containers *ContainersDefinition) GetState() *commonv1.State {
	return containers.State
}

func (containers *ContainersDefinition) SetState(state *commonv1.State) {
	containers.State = state
}

func (containers *ContainersDefinition) GetKind() string {
	return static.KIND_CONTAINERS
}

func (containers *ContainersDefinition) ResolveReferences(obj iobjects.ObjectInterface) ([]idefinitions.IDefinition, error) {
	return nil, nil
}

func (containers *ContainersDefinition) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, containers)
}

func (containers *ContainersDefinition) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(containers)
	return bytes, err
}

func (containers *ContainersDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(containers)
	return string(bytes), err
}

func (containers *ContainersDefinition) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(containers)
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
