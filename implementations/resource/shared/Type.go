package shared

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}
