package shared

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Registry platforms.Registry
	User     *authentication.User
	Watchers *watcher.Containers
	DnsCache *dns.Records
	Manager  *manager.Manager
	Client   *clients.Http
	Replay   bool
}

func (shared *Shared) GetCluster() *cluster.Cluster {
	return shared.Manager.Cluster
}
func (shared *Shared) Drain() {
	shared.Watchers.Drain()
}
func (shared *Shared) IsDrained() bool { return len(shared.Watchers.Watchers) == 0 }
