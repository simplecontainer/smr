package shared

import (
	"github.com/simplecontainer/smr/implementations/containers/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

type Shared struct {
	Watcher *watcher.ContainersWatcher
	Manager *manager.Manager
	Client  *http.Client
}
