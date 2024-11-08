package watcher

func (ContainersWatcher *ContainersWatcher) AddOrUpdate(groupidentifier string, container *Containers) {
	ContainersWatcher.Containers[groupidentifier] = container
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
