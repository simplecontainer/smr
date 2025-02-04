package shared

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/kinds/container/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Registry *registry.Registry
	User     *authentication.User
	Watchers *watcher.Containers
	DnsCache *dns.Records
	Manager  *manager.Manager
	Client   *client.Http
}
