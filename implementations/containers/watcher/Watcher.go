package watcher

func (ContainersWatcher *ContainersWatcher) AddOrUpdate(groupidentifier string, container *Containers) {
	if ContainersWatcher.Containers[groupidentifier] == nil {
		ContainersWatcher.Containers[groupidentifier] = container
	} else {
		ContainersWatcher.Containers[groupidentifier] = container
	}
}

func (ContainersWatcher *ContainersWatcher) Remove(groupidentifier string) bool {
	if ContainersWatcher.Containers[groupidentifier] == nil {
		return true
	} else {
		delete(ContainersWatcher.Containers, groupidentifier)
		return true
	}
}

func (ContainersWatcher *ContainersWatcher) Find(groupidentifier string) *Containers {
	if ContainersWatcher.Containers[groupidentifier] != nil {
		return ContainersWatcher.Containers[groupidentifier]
	} else {
		return nil
	}
}
