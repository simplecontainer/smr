package watcher

import (
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"sync"
)

func NewWatchers() *RepositoryWatcher {
	return &RepositoryWatcher{
		Repositories: map[string]*Gitops{},
		Lock:         &sync.RWMutex{},
	}
}

func (RepositoryWatcher *RepositoryWatcher) AddOrUpdate(groupidentifier string, gitopsWatcher *Gitops) {
	RepositoryWatcher.Lock.RLock()
	defer RepositoryWatcher.Lock.RUnlock()

	RepositoryWatcher.Repositories[groupidentifier] = gitopsWatcher
}

func (RepositoryWatcher *RepositoryWatcher) Remove(groupidentifier string) bool {
	RepositoryWatcher.Lock.Lock()
	defer RepositoryWatcher.Lock.Unlock()

	if RepositoryWatcher.Repositories[groupidentifier] == nil {
		return true
	} else {
		delete(RepositoryWatcher.Repositories, groupidentifier)
		return true
	}
}

func (RepositoryWatcher *RepositoryWatcher) Drain() {
	RepositoryWatcher.Lock.Lock()
	defer RepositoryWatcher.Lock.Unlock()

	for _, watcher := range RepositoryWatcher.Repositories {
		watcher.Gitops.GetStatus().SetState(status.PENDING_DELETE)
		watcher.GitopsQueue <- watcher.Gitops
	}
}

func (RepositoryWatcher *RepositoryWatcher) Find(groupidentifier string) *Gitops {
	RepositoryWatcher.Lock.RLock()
	defer RepositoryWatcher.Lock.RUnlock()

	if RepositoryWatcher.Repositories[groupidentifier] != nil {
		return RepositoryWatcher.Repositories[groupidentifier]
	} else {
		return nil
	}
}
