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
	Enabled bool              `json:"enabled"`
	Name    string            `validate:"required" json:"name"`
	Group   string            `validate:"required" json:"group"`
	Labels  map[string]string `json:"labels"`
}

type ContainerSpec struct {
	Container ContainerInternal `validate:"required" json:"container" `
}

type ContainerInternal struct {
	Image         string               `validate:"required" json:"image"`
	Tag           string               `validate:"required" json:"tag"`
	Envs          []string             `json:"envs"`
	Entrypoint    []string             `json:"entrypoint"`
	Command       []string             `json:"command"`
	Dependencies  []ContainerDependsOn `json:"dependencies"`
	Readiness     []ContainerReadiness `json:"readiness"`
	Networks      []string             `validate:"required" json:"networks"`
	Ports         []map[string]string  `json:"ports"`
	Volumes       []map[string]string  `json:"volumes"`
	Operators     []map[string]any     `json:"operators"`
	Configuration map[string]string    `json:"configuration"`
	Resources     []map[string]string  `json:"resources"`
	Replicas      int                  `validate:"required" json:"replicas"`
	Capabilities  []string             `json:"capabilities"`
	Privileged    bool                 `json:"privileged"`
	NetworkMode   string               `json:"network_mode"`
}

//type ContainerInternalV2 struct {
//	Image         string               `validate:"required" json:"image"`
//	Tag           string               `validate:"required" json:"tag"`
//	Replicas      int                  `validate:"required" json:"replicas"`
//	Entrypoint    []string             `json:"entrypoint"`
//	Command       []string             `json:"command"`
//	Dependencies  []ContainerDependsOn `json:"dependencies"`
//	Readiness     []ContainerReadiness `json:"readiness"`
//	Networks      []ContainerNetwork   `validate:"required" json:"networks"`
//	NetworkMode   string               `json:"network_mode"`
//	Ports         []ContainerPort      `json:"ports"`
//	Volumes       []ContainerVolume    `json:"volumes"`
//	Envs          []string             `json:"envs"`
//	Configuration map[string]string    `json:"configuration"`
//	Resources     []ContainerResource  `json:"resources"`
//	Capabilities  []string             `json:"capabilities"`
//	Privileged    bool                 `json:"privileged"`
//}

type ContainerDependsOn struct {
	Name    string `validate:"required" json:"name"`
	Group   string `validate:"required" json:"group"`
	Timeout string `validate:"required" json:"timeout"`
}

type ContainerReadiness struct {
	Name     string            `validate:"required" json:"name"`
	Operator string            `json:"operator"`
	Timeout  string            `validate:"required" json:"timeout"`
	Body     map[string]string `json:"body"`
}

//type ContainerNetwork struct {
//	Network string
//}
//
//type ContainerPort struct {
//	Container string `json:"container"`
//	Host      string `json:"host"`
//}
//
//type ContainerVolume struct {
//	Host   string `json:"host"`
//	Target string `json:"target"`
//}
//
//type ContainerResource struct {
//	Name       string
//	Group      string
//	Key        string
//	MountPoint string
//}

func (definition *ContainerDefinition) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (definition *ContainerDefinition) Validate() (bool, error) {
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
