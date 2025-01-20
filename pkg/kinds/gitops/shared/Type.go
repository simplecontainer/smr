package shared

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/registry"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Registry *registry.Registry
	Watcher  *watcher.RepositoryWatcher
	Manager  *manager.Manager
	Client   *client.Http
}
