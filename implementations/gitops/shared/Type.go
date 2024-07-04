package shared

import (
	"github.com/simplecontainer/smr/implementations/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

type Shared struct {
	Watcher *watcher.RepositoryWatcher
	Manager *manager.Manager
	Client  *http.Client
}
