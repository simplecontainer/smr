package resource

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
)

type Resource struct {
	Started bool
	Shared  *Shared
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

func (shared *Shared) GetManager() *manager.Manager {
	return shared.Manager
}

const KIND string = static.KIND_RESOURCE
