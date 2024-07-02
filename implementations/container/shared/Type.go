package shared

import (
	"github.com/qdnqn/smr/implementations/container/watcher"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/registry"
)

type Shared struct {
	Registry *registry.Registry
	Watcher  *watcher.ContainerWatcher
	DnsCache *dns.Records
	Manager  *manager.Manager
}
