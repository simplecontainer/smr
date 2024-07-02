package watcher

func (repositorywatcher *RepositoryWatcher) AddOrUpdate(groupidentifier string, gitops *Gitops) {
	if repositorywatcher.Repositories[groupidentifier] == nil {
		repositorywatcher.Repositories[groupidentifier] = gitops
	} else {
		repositorywatcher.Repositories[groupidentifier] = gitops
	}
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
