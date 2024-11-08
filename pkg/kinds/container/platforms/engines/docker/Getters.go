package docker

import (
	"fmt"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/static"
)

func (container *Docker) GetDefinition() v1.ContainerDefinition {
	return container.Definition
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

func (container *Docker) GetGroupIdentifier() string {
	return fmt.Sprintf("%s.%s", container.Group, container.GeneratedName)
}

func (container *Docker) GetDomain(network string) string {
	return fmt.Sprintf("%s.%s.%s.%s", network, container.Group, container.GeneratedName, static.SMR_LOCAL_DOMAIN)
}

func (container *Docker) GetHeadlessDomain(network string) string {
	return fmt.Sprintf("%s.%s.%s.%s", network, container.Group, container.Name, static.SMR_LOCAL_DOMAIN)
}
