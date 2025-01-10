package watcher

func (gitopsWatcher *Gitops) ShouldSync() bool {
	return gitopsWatcher.Gitops.AutomaticSync || gitopsWatcher.Gitops.ManualSync
}

func (gitopsWatcher *Gitops) Backoff() bool {
	if gitopsWatcher.BackOff.Failure > 5 {
		gitopsWatcher.BackOff.BackOff = true
	} else {
		gitopsWatcher.BackOff.Failure += 1
	}

	return gitopsWatcher.BackOff.BackOff
}
