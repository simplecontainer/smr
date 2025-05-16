package shared

import (
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/registry"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Registry *registry.Registry
	Watchers *watcher.RepositoryWatcher
	Manager  *manager.Manager
	Client   *clients.Http
}

func (shared *Shared) GetCluster() *cluster.Cluster {
	return shared.Manager.Cluster
}
func (shared *Shared) Drain() {
	shared.Watchers.Drain()
}
func (shared *Shared) IsDrained() bool { return len(shared.Watchers.Repositories) == 0 }
