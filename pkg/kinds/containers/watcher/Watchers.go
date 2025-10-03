package watcher

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"sync"
)

func NewWatchers() *Containers {
	return &Containers{
		Watchers: map[string]*Container{},
		Lock:     &sync.RWMutex{},
	}
}

func (ContainerWatcher *Containers) AddOrUpdate(groupidentifier string, container *Container) {
	ContainerWatcher.Lock.Lock() // Changed from RLock
	defer ContainerWatcher.Lock.Unlock()
	ContainerWatcher.Watchers[groupidentifier] = container
}

func (ContainerWatcher *Containers) Remove(groupidentifier string) {
	ContainerWatcher.Lock.Lock()
	defer ContainerWatcher.Lock.Unlock()

	delete(ContainerWatcher.Watchers, groupidentifier)
}

func (ContainerWatcher *Containers) Drain() {
	ContainerWatcher.Lock.Lock()
	defer ContainerWatcher.Lock.Unlock()

	for _, watcher := range ContainerWatcher.Watchers {
		if watcher.Container.GetDefinition().GetMeta().GetRuntime() != nil &&
			watcher.Container.GetDefinition().GetMeta().GetRuntime().GetOwner() != nil &&
			watcher.Container.GetDefinition().GetMeta().GetRuntime().GetOwner().IsEmpty() {
			watcher.Container.GetStatus().SetState(status.DELETE)

			select {
			case watcher.DeleteC <- watcher.Container:
			default:
				watcher.Logger.Warn("delete channel full, skipping")
			}
		}
	}
}

func (ContainerWatcher *Containers) Find(groupidentifier string) *Container {
	ContainerWatcher.Lock.RLock()
	defer ContainerWatcher.Lock.RUnlock()

	if ContainerWatcher.Watchers[groupidentifier] != nil {
		return ContainerWatcher.Watchers[groupidentifier]
	} else {
		return nil
	}
}

func (ContainerWatcher *Containers) ForEach(fn func(groupIdentifier string, watcher *Container)) {
	ContainerWatcher.Lock.RLock()
	snapshot := make(map[string]*Container, len(ContainerWatcher.Watchers))
	for k, v := range ContainerWatcher.Watchers {
		snapshot[k] = v
	}
	ContainerWatcher.Lock.RUnlock()

	for k, v := range snapshot {
		if v != nil && !v.Done {
			fn(k, v)
		}
	}
}

func (ContainerWatcher *Containers) GetSnapshot() []*Container {
	ContainerWatcher.Lock.RLock()
	defer ContainerWatcher.Lock.RUnlock()

	snapshot := make([]*Container, 0, len(ContainerWatcher.Watchers))
	for _, watcher := range ContainerWatcher.Watchers {
		if watcher != nil {
			snapshot = append(snapshot, watcher)
		}
	}
	return snapshot
}
