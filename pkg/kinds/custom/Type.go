package custom

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
)

type Custom struct {
	Started bool
	Shared  *Shared
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = static.KIND_CUSTOM
