package shared

import (
	"github.com/qdnqn/smr/implementations/containers/watcher"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/registry"
)

type Shared struct {
	Registry *registry.Registry
	Watcher  *watcher.ContainersWatcher
	Manager  *manager.Manager
}
