package volume

import (
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

type Volume struct {
	Started bool
	Shared  *Shared
	Client  *http.Client
}

type Shared struct {
	Manager *manager.Manager
	Client  *clients.Http
}

func (shared *Shared) GetCluster() *cluster.Cluster {
	return shared.Manager.Cluster
}
func (shared *Shared) Drain()          {}
func (shared *Shared) IsDrained() bool { return true }

const KIND string = static.KIND_VOLUME
