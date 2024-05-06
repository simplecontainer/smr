package gitops

func (repositorywatcher *RepositoryWatcher) AddOrUpdate(repository string, gitops *Gitops) {
	if repositorywatcher.Repositories[repository] == nil {
		repositorywatcher.Repositories[repository] = gitops
	} else {
		repositorywatcher.Repositories[repository] = gitops
	}
}

func (repositorywatcher *RepositoryWatcher) Remove(repository string) bool {
	if repositorywatcher.Repositories[repository] == nil {
		return true
	} else {
		delete(repositorywatcher.Repositories, repository)
		return true
	}
}

func (repositorywatcher *RepositoryWatcher) Find(repository string) *Gitops {
	if repositorywatcher.Repositories[repository] != nil {

		return repositorywatcher.Repositories[repository]

	} else {
		return nil
	}
}
