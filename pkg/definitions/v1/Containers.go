package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/static"
	"gopkg.in/yaml.v3"
)

type ContainersDefinition struct {
	Kind          string              `json:"kind" yaml:"kind" validate:"required"`
	Prefix        string              `json:"prefix" yaml:"prefix" validate:"required"`
	Meta          *commonv1.Meta      `json:"meta" yaml:"meta" validate:"required"`
	Spec          *ContainersInternal `json:"spec" yaml:"spec" validate:"required"`
	InitContainer *ContainersInternal `json:"initContainer,omitempty" yaml:"initContainer,omitempty"`
	State         *commonv1.State     `json:"state,omitempty" yaml:"state,omitempty"`
}

type ContainersInternal struct {
	Image          string                     `json:"image,omitempty" yaml:"image,omitempty" validate:"required"`
	Tag            string                     `json:"tag,omitempty" yaml:"tag,omitempty" validate:"required"`
	RegistryAuth   string                     `json:"registryAuth,omitempty" yaml:"registryAuth,omitempty"`
	Envs           []string                   `json:"envs,omitempty" yaml:"envs,omitempty"`
	Entrypoint     []string                   `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Args           []string                   `json:"args,omitempty" yaml:"args,omitempty"`
	Dependencies   []ContainersDependsOn      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Readiness      []ContainersReadiness      `json:"readiness,omitempty" yaml:"readiness,omitempty"`
	Networks       []ContainersNetwork        `json:"networks,omitempty" yaml:"networks,omitempty"`
	Ports          []ContainersPort           `json:"ports,omitempty" yaml:"ports,omitempty"`
	Volumes        []ContainersVolume         `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Configuration  map[string]string          `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	Resources      []ContainersResource       `json:"resources,omitempty" yaml:"resources,omitempty"`
	Configurations []ContainersConfigurations `json:"configurations,omitempty" yaml:"configurations,omitempty"`
	Replicas       uint64                     `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Capabilities   []string                   `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	User           string                     `json:"user,omitempty" yaml:"user,omitempty"`
	GroupAdd       []string                   `json:"groupAdd,omitempty" yaml:"groupAdd,omitempty"`
	Privileged     bool                       `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	NetworkMode    string                     `json:"network_mode,omitempty" yaml:"network_mode,omitempty"`
	Spread         *ContainersSpread          `json:"spread,omitempty" yaml:"spread,omitempty"`
	Nodes          []string                   `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Dns            []string                   `json:"dns,omitempty" yaml:"dns,omitempty"`
}

func NewContainers() *ContainersDefinition {
	return &ContainersDefinition{
		Kind:   "",
		Prefix: "",
		Meta: &commonv1.Meta{
			Group:   "",
			Name:    "",
			Labels:  nil,
			Runtime: &commonv1.Runtime{},
		},
		Spec:  &ContainersInternal{},
		State: nil,
	}
}

type ContainersRegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
	Spread string   `json:"spread,omitempty"`
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
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	HostPath    string    `json:"hostPath"`
	MountPoint  string    `json:"mountPoint"`
	SubPath     string    `json:"subPath" default:"/"`
	Permissions *FileInfo `json:"fileInfo"`
}

type ContainersResource struct {
	Name        string
	Group       string
	Key         string
	MountPoint  string
	Permissions *FileInfo
}

type FileInfo struct {
	Owner       *int    `json:"owner"`
	Group       *int    `json:"group"`
	Permissions *uint32 `json:"permissions"`
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

func (containers *ContainersDefinition) GetMeta() *commonv1.Meta {
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

func (containers *ContainersDefinition) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(containers)
	return bytes, err
}

func (containers *ContainersDefinition) ToYAML() ([]byte, error) {
	bytes, err := yaml.Marshal(containers)
	return bytes, err
}

func (containers *ContainersDefinition) ToJSONString() (string, error) {
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
