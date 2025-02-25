package reconcile

import (
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"go.uber.org/zap"
	"time"
)

func Gitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	if gitopsWatcher.Done {
		return
	}

	gitopsObj := gitopsWatcher.Gitops

	if gitopsObj.GetStatus().Reconciling {
		gitopsWatcher.Logger.Info("gitops already reconciling, waiting for the free slot")
		return
	}

	gitopsObj.GetStatus().Reconciling = true

	if gitopsObj.ForcePoll {
		gitopsObj.ForcePoll = false
		gitopsObj.GetStatus().SetState(status.STATUS_CLONING_GIT)
	}

	newState, reconcile := Reconcile(shared, gitopsWatcher)

	gitopsObj.GetStatus().Reconciling = false

	transitioned := gitopsObj.GetStatus().TransitionState(gitopsObj.GetGroup(), gitopsObj.GetName(), newState)

	if !transitioned {
		gitopsWatcher.Logger.Error("failed to transition state",
			zap.String("old", gitopsObj.GetStatus().State.State),
			zap.String("new", newState))
	}

	err := shared.Registry.Sync(gitopsObj.GetGroup(), gitopsObj.GetName())

	if err != nil {
		gitopsWatcher.Logger.Error(err.Error())
	}

	if reconcile {
		gitopsWatcher.GitopsQueue <- gitopsObj
	} else {
		DispatchEventChange(shared, gitopsWatcher.Gitops)

		switch gitopsObj.GetStatus().GetState() {
		case status.STATUS_DRIFTED:
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			return
		case status.STATUS_INSPECTING:
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			return
		default:
			if gitopsObj.GetStatus().GetCategory() == status.CATEGORY_END {
				gitopsWatcher.Ticker.Stop()
			} else {
				gitopsWatcher.Ticker.Reset(5 * time.Second)
			}
		}
	}
}
