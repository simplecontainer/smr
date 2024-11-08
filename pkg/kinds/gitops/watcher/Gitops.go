package watcher

func (gitops *Gitops) Backoff() bool {
	if gitops.BackOff.Failure > 5 {
		gitops.BackOff.BackOff = true
	} else {
		gitops.BackOff.Failure += 1
	}

	return gitops.BackOff.BackOff
}
