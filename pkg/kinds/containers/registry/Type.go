package registry

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"sync"
)

type Registry struct {
	Containers     map[string]platforms.IContainer
	ContainersLock sync.RWMutex
	BackOffTracker map[string]map[string]uint64
	Client         *clients.Http
	User           *authentication.User
}
