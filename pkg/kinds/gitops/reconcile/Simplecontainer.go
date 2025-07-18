package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/packer"
	"go.uber.org/zap"
)

type StateHandlerFunc func(shared *shared.Shared, gw *watcher.Gitops) (string, bool)

var stateHandlers = map[string]StateHandlerFunc{
	status.CREATED:             handleCreated,
	status.CLONING_GIT:         handleCloningGit,
	status.COMMIT_GIT:          handleCommitGit,
	status.CLONED_GIT:          handleClonedGit,
	status.INVALID_GIT:         handleInvalidGit,
	status.INVALID_DEFINITIONS: handleInvalidDefinitions,
	status.SYNCING:             handleSyncing,
	status.INSPECTING:          handleInspecting,
	status.SYNCING_STATE:       handleSyncingState,
	status.INSYNC:              handleInSync,
	status.DRIFTED:             handleDrifted,
	status.DELETE:              handleDelete,
}

func Reconcile(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	state := gw.Gitops.GetStatus().GetState()
	if handler, ok := stateHandlers[state]; ok {
		return handler(shared, gw)
	}
	return status.CREATED, true
}

func handleCreated(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Logger.Info(fmt.Sprintf("%s is created", gw.Gitops.GetName()))
	err := gw.Gitops.Prepare(shared.Client, gw.User)
	if err != nil {
		gw.Logger.Error(fmt.Sprintf("%s failed to resolve gitops references and generate auth credentials", gw.Gitops.GetName()))
		gw.Logger.Error(err.Error())
		return status.INVALID_GIT, true
	}
	return status.CLONING_GIT, true
}

func handleCloningGit(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	headRemote, err := gw.Gitops.GetGit().RemoteHead()
	if err != nil {
		gw.Logger.Error(err.Error())
		return status.INVALID_GIT, true
	}
	if headRemote.IsZero() || gw.Gitops.GetCommit().ID() != headRemote {
		gw.Logger.Info("found new commit on remote - pulling latest")

		err = gw.Gitops.SetCommit(gw.Gitops.GetGit().Fetch())
		if err != nil {
			gw.Logger.Error(err.Error())
			return status.INVALID_GIT, true
		}

		return status.CLONED_GIT, true
	}

	return status.CLONED_GIT, true
}

func handleCommitGit(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	if !gw.Gitops.GetQueue().IsEmpty() {
		err := gw.Gitops.Commit(gw.Logger, shared.Client, gw.User, gw.Gitops.GetQueue().Pop())

		if err != nil {
			gw.Logger.Error(err.Error())
			return status.INVALID_GIT, true
		} else {
			if !gw.Gitops.GetQueue().IsEmpty() {
				return status.COMMIT_GIT, true
			} else {
				gw.Gitops.SetForceClone(true)
				return status.CLONING_GIT, true
			}
		}
	} else {
		return status.CLONING_GIT, true
	}
}

func handleClonedGit(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	var err error

	if len(gw.Gitops.GetPack().Definitions) == 0 {
		err = gw.Gitops.SetPack(packer.Read(fmt.Sprintf("%s/%s", gw.Gitops.GetGit().Directory, gw.Gitops.GetDirectory()), nil, shared.Manager.Kinds))
		if err != nil {
			return status.INVALID_DEFINITIONS, true
		}
	} else {
		tmp, err := packer.Read(fmt.Sprintf("%s/%s", gw.Gitops.GetGit().Directory, gw.Gitops.GetDirectory()), nil, shared.Manager.Kinds)
		if err != nil {
			return status.INVALID_DEFINITIONS, true
		}

		err = gw.Gitops.Update(tmp)
		if err != nil {
			return status.INVALID_DEFINITIONS, true
		}
	}
	return status.INSPECTING, true
}

func handleInvalidGit(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Logger.Info("git configuration is invalid or pull failed")
	return status.INVALID_GIT, false
}

func handleInvalidDefinitions(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Logger.Info("definitions are invalid")
	return status.INVALID_DEFINITIONS, false
}

func handleSyncing(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Gitops.GetStatus().GetPending().Set(status.PENDING_SYNC)
	gw.Logger.Info(fmt.Sprintf("attempt to sync commit %s", gw.Gitops.GetCommit().ID()))

	if len(gw.Gitops.GetPack().Definitions) == 0 {
		gw.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gw.Gitops.GetGit().Directory, gw.Gitops.GetDirectory()))
		return status.INVALID_DEFINITIONS, true
	}

	errs := []error{}
	_, errs = gw.Gitops.Sync(gw.Logger, shared.Client, gw.User)
	if len(errs) > 0 {
		for _, e := range errs {
			gw.Logger.Error(e.Error())
		}
		return status.INVALID_DEFINITIONS, true
	}

	gw.Gitops.GetStatus().LastSyncedCommit = gw.Gitops.GetCommit().ID()
	gw.Gitops.GetStatus().InSync = true
	gw.Gitops.SetForceSync(false)
	gw.Logger.Info(fmt.Sprintf("commit %s synced", gw.Gitops.GetStatus().LastSyncedCommit))
	return status.INSPECTING, true
}

func handleInspecting(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	drifted, errs := gw.Gitops.Drift(shared.Client, gw.User)
	if len(errs) > 0 {
		for _, e := range errs {
			gw.Logger.Error(e.Error())
		}
		return status.INVALID_DEFINITIONS, true
	}
	if gw.Gitops.GetStatus().InSync {
		gw.Gitops.GetStatus().InSync = !drifted
	}
	return status.SYNCING_STATE, true
}

func handleSyncingState(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Gitops.GetStatus().GetPending().Set(status.PENDING_SYNC)
	gw.Logger.Info("attempt to sync state")

	if len(gw.Gitops.GetPack().Definitions) == 0 {
		gw.Logger.Error(fmt.Sprintf("no valid definitions found: %s/%s", gw.Gitops.GetGit().Directory, gw.Gitops.GetDirectory()))
		return status.INVALID_DEFINITIONS, true
	}

	errs := []error{}
	_, errs = gw.Gitops.SyncState(gw.Logger, shared.Client, gw.User)

	if len(errs) > 0 {
		for _, e := range errs {
			gw.Logger.Error(e.Error())
		}
		return status.INVALID_DEFINITIONS, true
	}

	gw.Logger.Info("state synced")

	if gw.Gitops.GetStatus().InSync {
		return status.INSYNC, true
	}

	return status.DRIFTED, true
}

func handleInSync(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	return status.INSYNC, false
}

func handleDrifted(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Logger.Info("drift detected")
	if gw.Gitops.ShouldSync() {
		return status.SYNCING, true
	}
	return status.DRIFTED, false
}

func handleDelete(shared *shared.Shared, gw *watcher.Gitops) (string, bool) {
	gw.Logger.Info("triggering context cancel")
	err := gw.Gitops.GetStatus().GetPending().Set(status.PENDING_DELETE)
	if err != nil {
		logger.Log.Error("failed to set pending delete state", zap.Error(err))
	}
	return "", false
}
