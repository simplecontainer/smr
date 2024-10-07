package shared

import (
	"github.com/simplecontainer/smr/implementations/containers/watcher"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Watcher *watcher.ContainersWatcher
	Manager *manager.Manager
	Client  *client.Http
}
