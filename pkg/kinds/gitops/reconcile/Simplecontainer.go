package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/packer"
)

func Reconcile(shared *shared.Shared, gitopsWatcher *watcher.Gitops) (string, bool) {
	gitopsObj := gitopsWatcher.Gitops

	switch gitopsWatcher.Gitops.Status.GetState() {
	case status.CREATED:
		gitopsWatcher.Logger.Info(fmt.Sprintf("%s is created", gitopsObj.GetName()))

		err := gitopsWatcher.Gitops.Prepare(shared.Client, gitopsWatcher.User)

		if err != nil {
			gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to resolve gitops references and generate auth credentials", gitopsObj.GetName()))
			gitopsWatcher.Logger.Error(err.Error())

			return status.INVALID_GIT, true
		}

		return status.CLONING_GIT, true
	case status.CLONING_GIT:
		headRemote, err := gitopsObj.Git.RemoteHead()
		gitopsObj.ForcePoll = false

		if err != nil {
			gitopsWatcher.Logger.Error(err.Error())
			return status.INVALID_GIT, true
		} else {
			if headRemote.IsZero() || gitopsObj.Commit.ID() != headRemote {
				gitopsWatcher.Logger.Info(fmt.Sprintf("found new commit on remote - pulling latest"))
				gitopsObj.Commit, err = gitopsWatcher.Gitops.Git.Fetch()

				if err != nil {
					gitopsWatcher.Logger.Error(err.Error())
					return status.INVALID_GIT, true
				} else {
					return status.CLONED_GIT, true
				}
			} else {
				return status.CLONED_GIT, true
			}
		}
	case status.CLONED_GIT:
		var err error

		if len(gitopsObj.Definitions) == 0 {
			gitopsObj.Definitions, err = packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)

			if err != nil {
				return status.INVALID_DEFINITIONS, true
			}
		} else {
			var tmp []*common.Request
			tmp, err = packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)

			if err != nil {
				return status.INVALID_DEFINITIONS, true
			}

			err = gitopsObj.Update(tmp)

			if err != nil {
				return status.INVALID_DEFINITIONS, true
			}
		}

		return status.INSPECTING, true
	case status.INVALID_GIT:
		gitopsWatcher.Logger.Info("git configuration is invalid or pull failed")
		return status.INVALID_GIT, false
	case status.INVALID_DEFINITIONS:
		gitopsWatcher.Logger.Info("definitions are invalid")
		return status.INVALID_DEFINITIONS, false
	case status.SYNCING:
		gitopsWatcher.Gitops.Status.GetPending().Set(status.PENDING_SYNC)
		gitopsWatcher.Logger.Info(fmt.Sprintf("attempt to sync commit %s", gitopsWatcher.Gitops.Commit.ID()))

		if len(gitopsObj.Definitions) == 0 {
			gitopsWatcher.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath))
			return status.INVALID_DEFINITIONS, true
		} else {
			var errs []error
			_, errs = gitopsWatcher.Gitops.Sync(gitopsWatcher.Logger, shared.Client, gitopsWatcher.User)

			if len(errs) > 0 {
				for _, e := range errs {
					gitopsWatcher.Logger.Error(e.Error())
				}

				return status.INVALID_DEFINITIONS, true
			}

			gitopsObj.GetStatus().LastSyncedCommit = gitopsObj.Commit.ID()
			gitopsObj.GetStatus().InSync = true
			gitopsObj.ForceSync = false

			gitopsWatcher.Logger.Info(fmt.Sprintf("commit %s synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
			return status.INSYNC, true
		}

	case status.INSPECTING:
		var errs []error
		var drifted bool

		drifted, errs = gitopsWatcher.Gitops.Drift(shared.Client, gitopsWatcher.User)

		if len(errs) > 0 {
			for _, e := range errs {
				gitopsWatcher.Logger.Error(e.Error())
			}

			return status.INVALID_DEFINITIONS, true
		}

		if gitopsObj.GetStatus().InSync {
			gitopsObj.GetStatus().InSync = drifted == false
		}

		return status.SYNCING_STATE, true
	case status.SYNCING_STATE:
		gitopsWatcher.Gitops.Status.GetPending().Set(status.PENDING_SYNC)
		gitopsWatcher.Logger.Info(fmt.Sprintf("attempt to sync state"))

		if len(gitopsObj.Definitions) == 0 {
			gitopsWatcher.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath))
			return status.INVALID_DEFINITIONS, true
		} else {
			var errs []error
			_, errs = gitopsWatcher.Gitops.SyncState(gitopsWatcher.Logger, shared.Client, gitopsWatcher.User)

			if len(errs) > 0 {
				for _, e := range errs {
					gitopsWatcher.Logger.Error(e.Error())
				}

				return status.INVALID_DEFINITIONS, true
			}

			gitopsWatcher.Logger.Info(fmt.Sprintf("state synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
		}

		if gitopsObj.GetStatus().InSync {
			return status.INSYNC, true
		} else {
			return status.DRIFTED, true
		}
	case status.INSYNC:
		return status.INSYNC, false
	case status.DRIFTED:
		gitopsWatcher.Logger.Info("drift detected")

		if gitopsWatcher.Gitops.ShouldSync() {
			return status.SYNCING, true
		} else {
			return status.DRIFTED, false
		}
	case status.PENDING_DELETE:
		gitopsWatcher.Logger.Info("triggering context cancel")
		gitopsWatcher.Gitops.Status.GetPending().Set(status.PENDING_DELETE)

		return "", false
	default:
		return status.CREATED, true
	}
}
