package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/packer"
)

func Reconcile(shared *shared.Shared, gitopsWatcher *watcher.Gitops) (string, bool) {
	gitopsObj := gitopsWatcher.Gitops

	switch gitopsWatcher.Gitops.Status.GetState() {
	case status.STATUS_CREATED:
		gitopsWatcher.Logger.Info(fmt.Sprintf("%s is created", gitopsObj.GetName()))

		err := gitopsWatcher.Gitops.Prepare(shared.Client, gitopsWatcher.User)

		if err != nil {
			gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to resolve gitops references and generate auth credentials", gitopsObj.GetName()))
			gitopsWatcher.Logger.Error(err.Error())

			return status.STATUS_INVALID_GIT, true
		}

		return status.STATUS_CLONING_GIT, true
	case status.STATUS_CLONING_GIT:
		headRemote, err := gitopsObj.Git.RemoteHead()
		gitopsObj.ForcePoll = false

		if err != nil {
			gitopsWatcher.Logger.Error(err.Error())
			return status.STATUS_INVALID_GIT, true
		} else {
			if headRemote.IsZero() || gitopsObj.Commit.ID() != headRemote {
				gitopsWatcher.Logger.Info(fmt.Sprintf("found new commit on remote - pulling latest"))
				gitopsObj.Commit, err = gitopsWatcher.Gitops.Git.Fetch()

				if err != nil {
					gitopsWatcher.Logger.Error(err.Error())
					return status.STATUS_INVALID_GIT, true
				} else {
					return status.STATUS_CLONED_GIT, true
				}
			} else {
				return status.STATUS_CLONED_GIT, true
			}
		}
	case status.STATUS_CLONED_GIT:
		if gitopsWatcher.Gitops.ShouldSync() {
			return status.STATUS_SYNCING, true
		} else {
			return status.STATUS_INSPECTING, true
		}
	case status.STATUS_INVALID_GIT:
		gitopsWatcher.Logger.Info("git configuration is invalid or pull failed")
		return status.STATUS_INVALID_GIT, false
	case status.STATUS_INVALID_DEFINITIONS:
		gitopsWatcher.Logger.Info("definitions are invalid")
		return status.STATUS_INVALID_DEFINITIONS, false
	case status.STATUS_SYNCING:
		gitopsWatcher.Logger.Info(fmt.Sprintf("attempt to sync commit %s", gitopsWatcher.Gitops.Commit.ID()))

		if gitopsObj.Status.LastSyncedCommit != gitopsObj.Commit.ID() || !gitopsObj.Status.InSync {
			defs, err := packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)

			if err != nil {
				gitopsWatcher.Logger.Error(err.Error())
				return status.STATUS_INVALID_DEFINITIONS, true
			} else {
				if len(defs) == 0 {
					gitopsWatcher.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath))
					return status.STATUS_INVALID_DEFINITIONS, true
				} else {
					gitopsWatcher.Children, err = gitopsWatcher.Gitops.Sync(gitopsWatcher.Logger, shared.Client, gitopsWatcher.User, defs)

					if err != nil {
						gitopsWatcher.Logger.Error(err.Error())
						return status.STATUS_INVALID_DEFINITIONS, true
					}

					gitopsObj.GetStatus().LastSyncedCommit = gitopsObj.Commit.ID()
					gitopsObj.GetStatus().InSync = true
					gitopsObj.DoSync = false

					gitopsWatcher.Logger.Info(fmt.Sprintf("commit %s synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
					return status.STATUS_INSYNC, true
				}
			}
		} else {
			gitopsWatcher.Logger.Info("everything synced")
			return status.STATUS_INSYNC, true
		}
	case status.STATUS_INSPECTING:
		defs, err := packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)

		var drifted bool
		drifted, err = gitopsWatcher.Gitops.Drift(shared.Client, gitopsWatcher.User, defs)

		if err != nil {
			gitopsWatcher.Logger.Error(err.Error())
			return status.STATUS_INVALID_DEFINITIONS, true
		}

		gitopsObj.GetStatus().InSync = drifted == false

		if gitopsObj.GetStatus().InSync {
			return status.STATUS_INSYNC, true
		} else {
			return status.STATUS_DRIFTED, true
		}
	case status.STATUS_INSYNC:
		if gitopsObj.GetStatus().PreviousState.State == status.STATUS_INSYNC {
			return status.STATUS_INSPECTING, true
		} else {
			return status.STATUS_INSYNC, false
		}
	case status.STATUS_DRIFTED:
		if gitopsWatcher.Gitops.AutomaticSync {
			gitopsWatcher.Logger.Info("drift detected")
			return status.STATUS_SYNCING, true
		} else {
			return status.STATUS_DRIFTED, false
		}
	case status.STATUS_PENDING_DELETE:
		gitopsWatcher.Logger.Info("delete is in process")
		return status.STATUS_PENDING_DELETE, false
	default:
		return status.STATUS_CREATED, true
	}
}
