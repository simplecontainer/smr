package docker

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker/internal"
)

type Docker struct {
	DockerID      string
	DockerState   string
	Name          string
	GeneratedName string
	Labels        map[string]string
	Group         string
	Image         string
	Tag           string
	Replicas      uint64
	Env           []string
	Entrypoint    []string
	Args          []string
	Privileged    bool
	NetworkMode   string
	Networks      *internal.Networks
	Ports         *internal.Ports
	Volumes       *internal.Volumes
	Readiness     *internal.Readinesses
	Resources     *internal.Resources
	Capabilities  []string
	Definition    v1.ContainerDefinition
	Auth          string
}
