package shared

import (
	"github.com/simplecontainer/smr/implementations/common/interfaces"
	"github.com/simplecontainer/smr/implementations/container/registry"
	"github.com/simplecontainer/smr/implementations/container/watcher"
	"github.com/simplecontainer/smr/pkg/dns"
	"net/http"
)

type Shared struct {
	Registry *registry.Registry
	Watcher  *watcher.ContainerWatcher
	DnsCache *dns.Records
	Manager  interfaces.ManagerInterface
	Client   *http.Client
}
