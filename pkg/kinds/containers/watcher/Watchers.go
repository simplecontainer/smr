package watcher

func (ContainerWatcher *Containers) AddOrUpdate(groupidentifier string, container *Container) {
	ContainerWatcher.Lock.RLock()
	defer ContainerWatcher.Lock.RUnlock()
	ContainerWatcher.Watchers[groupidentifier] = container
}

func (ContainerWatcher *Containers) Remove(groupidentifier string) {
	ContainerWatcher.Lock.Lock()
	defer ContainerWatcher.Lock.Unlock()

	delete(ContainerWatcher.Watchers, groupidentifier)
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
