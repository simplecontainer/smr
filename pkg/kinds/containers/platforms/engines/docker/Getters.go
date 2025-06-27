package docker

import (
	"fmt"
	TDTypes "github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"strconv"
	"strings"
)

func (container *Docker) GetNetwork() map[string]net.IP {
	ips := make(map[string]net.IP)

	for _, network := range container.Networks.Networks {
		ips[network.Reference.Name] = net.ParseIP(network.Docker.IP)
	}

	return ips
}

func (container *Docker) GetReadiness() []*readiness.Readiness {
	return container.Readiness.Readinesses
}

func (container *Docker) GetState() (state.State, error) {
	dockerContainer, err := internal.Get(container.GeneratedName)

	if err != nil {
		container.DockerState = ""
		return state.State{}, err
	}

	container.DockerID = dockerContainer.ID
	container.DockerState = dockerContainer.State

	var inspected TDTypes.ContainerJSON
	inspected, err = internal.Inspect(container.DockerID)

	if err != nil {
		return state.State{}, err
	}

	return state.State{
		State: dockerContainer.State,
		Error: inspected.State.Error,
	}, nil
}

func (container *Docker) Get() (*TDTypes.Container, error) {
	dockerContainer, err := internal.Get(container.GeneratedName)

	if err != nil {
		return nil, err
	}

	container.DockerID = dockerContainer.ID
	container.DockerState = dockerContainer.State

	if dockerContainer.State == "running" {
		err = container.SyncNetwork()

		if err != nil {
			return nil, err
		}
	}

	return &dockerContainer, nil
}

func (container *Docker) GetId() string {
	return container.DockerID
}

func (container *Docker) GetDefinition() idefinitions.IDefinition {
	return &container.Definition
}

func (container *Docker) GetGeneratedName() string {
	return container.GeneratedName
}

func (container *Docker) GetName() string {
	return container.Name
}

func (container *Docker) GetGroup() string {
	return container.Group
}

func (container *Docker) GetIndex() (uint64, error) {
	tmp := strings.Split(container.GetGeneratedName(), "-")
	return strconv.ParseUint(tmp[len(tmp)-1], 10, 64)
}

func (container *Docker) GetGroupIdentifier() string {
	return fmt.Sprintf("%s/%s", container.Group, container.GeneratedName)
}

func (container *Docker) GetImageWithTag() string {
	return fmt.Sprintf("%s:%s", container.Image, container.Tag)
}

func (container *Docker) GetDomain(network string) string {
	return fmt.Sprintf("%s.%s.%s", network, container.GeneratedName, static.SMR_LOCAL_DOMAIN)
}

func (container *Docker) GetHeadlessDomain(network string) string {
	return fmt.Sprintf("%s.%s.%s.%s", network, container.Group, container.Name, static.SMR_LOCAL_DOMAIN)
}

func (container *Docker) GetInit() platforms.IPlatform {
	return container.Init
}

func (container *Docker) GetInitDefinition() v1.ContainersInternal {
	return container.Definition.InitContainer
}
