package docker

import (
	TDVolume "github.com/docker/docker/api/types/volume"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/smaps"
	"sync"
)

type Docker struct {
	Init           *Docker
	DockerID       string
	DockerState    string
	Name           string
	GeneratedName  string
	Labels         *internal.Labels
	Group          string
	Image          string
	Tag            string
	Replicas       uint64
	Configuration  *smaps.Smap
	Lock           sync.RWMutex `json:"-"`
	Env            []string
	Entrypoint     []string
	Args           []string
	Privileged     bool
	NetworkMode    string
	Networks       *internal.Networks
	Ports          *internal.Ports
	Volumes        *internal.Volumes
	VolumeInternal TDVolume.Volume `json:"-"`
	Readiness      *internal.Readinesses
	Resources      *internal.Resources
	Configurations *internal.Configurations
	Capabilities   []string
	Definition     v1.ContainersDefinition
	RegistryAuth   string
	Docker         DockerInternal
}

type DockerInternal struct {
	DNS []string
}
