package definitions

import "smr/pkg/network"

type Containers struct {
	Kind       string               `yaml:"kind"`
	Containers map[string]Container `mapstructure:"container"`
}

type Container struct {
	Meta Meta `mapstructure:"meta"`
	Spec Spec `mapstructure:"spec"`
}

type Meta struct {
	Enabled bool              `yaml:"enabled"`
	Name    string            `yaml:"name"`
	Group   string            `yaml:"group"`
	Labels  map[string]string `mapstructure:"labels"`
}

type Spec struct {
	Container ContainerInternal `mapstructure:"container"`
}

type ContainerInternal struct {
	Image         string                 `yaml:"image"`
	Tag           string                 `yaml:"tag""`
	Envs          []string               `yaml:"envs"`
	Entrypoint    []string               `yaml:"entrypoint"`
	Command       []string               `json:"command"`
	Dependencies  []DependsOn            `yaml:"dependencies"`
	Networks      []string               `yaml:"networks"`
	Ports         []network.PortMappings `yaml:"ports"`
	Volumes       []map[string]string    `yaml:"volumes"`
	Operators     []map[string]any       `yaml:"operators"`
	Configuration map[string]any         `mapstructure:"configuration"`
	Resources     []map[string]any       `mapstructure:"resources"`
	Replicas      int                    `yaml:"replicas"`
	Capabilities  []string               `json:"capabilities"`
	Privileged    bool                   `json:"privileged"`
	NetworkMode   string                 `json:"network_mode"`
}

type DependsOn struct {
	Name     string         `yaml:"name"`
	Operator string         `yaml:"operator"`
	Timeout  string         `yaml:"timeout"`
	Body     map[string]any `mapstructure:"body"`
	Solved   bool
}
