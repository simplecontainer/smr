package watcher

func (repositorywatcher *RepositoryWatcher) AddOrUpdate(groupidentifier string, gitopsWatcher *Gitops) {
	repositorywatcher.Repositories[groupidentifier] = gitopsWatcher
}

func (repositorywatcher *RepositoryWatcher) Remove(groupidentifier string) bool {
	if repositorywatcher.Repositories[groupidentifier] == nil {
		return true
	} else {
		delete(repositorywatcher.Repositories, groupidentifier)
		return true
	}
}

func (repositorywatcher *RepositoryWatcher) Find(groupidentifier string) *Gitops {
	if repositorywatcher.Repositories[groupidentifier] != nil {
		return repositorywatcher.Repositories[groupidentifier]
	} else {
		return nil
	}
}
