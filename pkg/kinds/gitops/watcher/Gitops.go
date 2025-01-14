package watcher

func (gitopsWatcher *Gitops) ShouldSync() bool {
	return gitopsWatcher.Gitops.AutomaticSync || gitopsWatcher.Gitops.ManualSync
}
