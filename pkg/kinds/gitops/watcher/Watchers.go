package watcher

func (RepositoryWatcher *RepositoryWatcher) AddOrUpdate(groupidentifier string, gitopsWatcher *Gitops) {
	RepositoryWatcher.Lock.RLock()
	RepositoryWatcher.Lock.RUnlock()

	RepositoryWatcher.Repositories[groupidentifier] = gitopsWatcher
}

func (RepositoryWatcher *RepositoryWatcher) Remove(groupidentifier string) bool {
	RepositoryWatcher.Lock.Lock()
	RepositoryWatcher.Lock.Unlock()

	if RepositoryWatcher.Repositories[groupidentifier] == nil {
		return true
	} else {
		delete(RepositoryWatcher.Repositories, groupidentifier)
		return true
	}
}

func (RepositoryWatcher *RepositoryWatcher) Find(groupidentifier string) *Gitops {
	RepositoryWatcher.Lock.RLock()
	RepositoryWatcher.Lock.RUnlock()

	if RepositoryWatcher.Repositories[groupidentifier] != nil {
		return RepositoryWatcher.Repositories[groupidentifier]
	} else {
		return nil
	}
}
