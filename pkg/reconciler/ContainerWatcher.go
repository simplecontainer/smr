package reconciler

func (ContainerWatcher *ContainerWatcher) AddOrUpdate(groupidentifier string, container *Container) {
	if ContainerWatcher.Container[groupidentifier] == nil {
		ContainerWatcher.Container[groupidentifier] = container
	} else {
		ContainerWatcher.Container[groupidentifier] = container
	}
}

func (ContainerWatcher *ContainerWatcher) Remove(groupidentifier string) bool {
	if ContainerWatcher.Container[groupidentifier] == nil {
		return true
	} else {
		delete(ContainerWatcher.Container, groupidentifier)
		return true
	}
}

func (ContainerWatcher *ContainerWatcher) Find(groupidentifier string) *Container {
	if ContainerWatcher.Container[groupidentifier] != nil {
		return ContainerWatcher.Container[groupidentifier]
	} else {
		return nil
	}
}
