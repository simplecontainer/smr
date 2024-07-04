package shared

import (
	"github.com/simplecontainer/smr/implementations/container/registry"
	"github.com/simplecontainer/smr/implementations/container/watcher"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

type Shared struct {
	Registry *registry.Registry
	Watcher  *watcher.ContainerWatcher
	DnsCache *dns.Records
	Manager  *manager.Manager
	Client   *http.Client
}
