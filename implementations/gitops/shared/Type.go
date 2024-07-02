package shared

import (
	"github.com/qdnqn/smr/implementations/gitops/watcher"
	"github.com/qdnqn/smr/pkg/manager"
)

type Shared struct {
	Watcher *watcher.RepositoryWatcher
	Manager *manager.Manager
}
