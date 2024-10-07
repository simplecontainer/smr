package shared

import (
	"github.com/simplecontainer/smr/implementations/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Watcher *watcher.RepositoryWatcher
	Manager *manager.Manager
	Client  *client.Http
}
