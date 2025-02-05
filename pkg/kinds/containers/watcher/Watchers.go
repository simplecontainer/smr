package watcher

func (ContainerWatcher *Containers) AddOrUpdate(groupidentifier string, container *Container) {
	ContainerWatcher.Watchers[groupidentifier] = container
}

func (ContainerWatcher *Containers) Remove(groupidentifier string) {
	delete(ContainerWatcher.Watchers, groupidentifier)
}

func (ContainerWatcher *Containers) Find(groupidentifier string) *Container {
	if ContainerWatcher.Watchers[groupidentifier] != nil {
		return ContainerWatcher.Watchers[groupidentifier]
	} else {
		return nil
	}
}
