package shared

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
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
	Client   *client.Http
	Replay   bool
}

func (shared *Shared) GetManager() *manager.Manager {
	return shared.Manager
}
