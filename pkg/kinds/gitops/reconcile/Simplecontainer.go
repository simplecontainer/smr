package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/static"
)

func DispatchEventInspect(shared *shared.Shared, gitopsObj *implementation.Gitops) {
	if gitopsObj.Definition.GetRuntime().GetOwner().Kind == static.KIND_GITOPS {
		event := events.New(events.EVENT_INSPECT, static.KIND_GITOPS, static.KIND_GITOPS, gitopsObj.Definition.GetRuntime().GetOwner().Group, gitopsObj.Definition.GetRuntime().GetOwner().Name, nil)

		if shared.Manager.Cluster.Node.NodeID == gitopsObj.Definition.GetRuntime().GetNode() {
			err := event.Propose(shared.Manager.Cluster.KVStore, gitopsObj.Definition.GetRuntime().GetNode())

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}
}

func DispatchEventDelete(shared *shared.Shared, gitopsObj *implementation.Gitops) {
	event := events.New(events.EVENT_DELETED, static.KIND_GITOPS, static.KIND_GITOPS, gitopsObj.Definition.GetMeta().Group, gitopsObj.Definition.GetMeta().Name, nil)

	if shared.Manager.Cluster.Node.NodeID == gitopsObj.Definition.GetRuntime().GetNode() {
		err := event.Propose(shared.Manager.Cluster.KVStore, gitopsObj.Definition.GetRuntime().GetNode())

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func DispatchEventChange(shared *shared.Shared, gitopsObj *implementation.Gitops) {
	event := events.New(events.EVENT_CHANGED, static.KIND_GITOPS, static.KIND_GITOPS, gitopsObj.Definition.GetMeta().Group, gitopsObj.Definition.GetMeta().Name, nil)

	if shared.Manager.Cluster.Node.NodeID == gitopsObj.Definition.GetRuntime().GetNode() {
		err := event.Propose(shared.Manager.Cluster.KVStore, gitopsObj.Definition.GetRuntime().GetNode())

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

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
		var err error

		if len(gitopsObj.Definitions) == 0 {
			gitopsObj.Definitions, err = packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)
		} else {
			var tmp []*common.Request
			tmp, err = packer.Read(fmt.Sprintf("%s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath), shared.Manager.Kinds)

			if err != nil {
				return status.STATUS_INVALID_DEFINITIONS, true
			}

			err = gitopsObj.Update(tmp)
		}

		if err != nil {
			return status.STATUS_INVALID_DEFINITIONS, true
		} else {
			if gitopsWatcher.Gitops.ShouldSync() {
				return status.STATUS_SYNCING, true
			} else {
				return status.STATUS_INSPECTING, true
			}
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
			if len(gitopsObj.Definitions) == 0 {
				gitopsWatcher.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gitopsObj.Git.Directory, gitopsObj.DirectoryPath))
				return status.STATUS_INVALID_DEFINITIONS, true
			} else {
				var errs []error
				_, errs = gitopsWatcher.Gitops.Sync(gitopsWatcher.Logger, shared.Client, gitopsWatcher.User)

				if len(errs) > 0 {
					for _, e := range errs {
						gitopsWatcher.Logger.Error(e.Error())
					}

					return status.STATUS_INVALID_DEFINITIONS, true
				}

				gitopsObj.GetStatus().LastSyncedCommit = gitopsObj.Commit.ID()
				gitopsObj.GetStatus().InSync = true
				gitopsObj.DoSync = false

				gitopsWatcher.Logger.Info(fmt.Sprintf("commit %s synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
				return status.STATUS_INSYNC, true
			}
		} else {
			gitopsWatcher.Logger.Info("everything synced")
			return status.STATUS_INSYNC, true
		}
	case status.STATUS_INSPECTING:
		var errs []error
		var drifted bool

		drifted, errs = gitopsWatcher.Gitops.Drift(shared.Client, gitopsWatcher.User)

		if len(errs) > 0 {
			for _, e := range errs {
				gitopsWatcher.Logger.Error(e.Error())
			}

			return status.STATUS_INVALID_DEFINITIONS, true
		}

		if gitopsObj.GetStatus().InSync {
			gitopsObj.GetStatus().InSync = drifted == false
		}

		if gitopsObj.GetStatus().InSync {
			return status.STATUS_INSYNC, true
		} else {
			return status.STATUS_DRIFTED, true
		}
	case status.STATUS_INSYNC:
		return status.STATUS_INSYNC, false
	case status.STATUS_DRIFTED:
		if gitopsWatcher.Gitops.AutomaticSync {
			gitopsWatcher.Logger.Info("drift detected")
			return status.STATUS_SYNCING, true
		} else {
			if gitopsObj.GetStatus().PreviousState.State == status.STATUS_DRIFTED {
				return status.STATUS_INSPECTING, true
			} else {
				return status.STATUS_DRIFTED, false
			}
		}
	case status.STATUS_PENDING_DELETE:
		gitopsWatcher.Logger.Info("triggering context cancel")

		// disable auto sync because child deletion
		gitopsWatcher.Gitops.AutomaticSync = false
		gitopsWatcher.Cancel()
		return status.STATUS_PENDING_DELETE, false
	default:
		return status.STATUS_CREATED, true
	}
}
