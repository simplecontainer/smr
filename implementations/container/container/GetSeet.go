package container

import (
	"fmt"
	"github.com/qdnqn/smr/pkg/static"
)

func (container *Container) GetDomain() string {
	return fmt.Sprintf("%s.%s.%s.", container.Static.Group, container.Static.GeneratedName, static.SMR_LOCAL_DOMAIN)
}

func (container *Container) GetHeadlessDomain() string {
	return fmt.Sprintf("%s.%s.%s.", container.Static.Group, container.Static.Name, static.SMR_LOCAL_DOMAIN)
}

func (container *Container) GetGroupIdentifier() string {
	return fmt.Sprintf("%s.%s", container.Static.Group, container.Static.GeneratedName)
}
