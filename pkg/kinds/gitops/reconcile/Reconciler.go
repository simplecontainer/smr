package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"go.uber.org/zap"
	"time"
)

func Gitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	if gitopsWatcher.Done || gitopsWatcher.Gitops.Status.PendingDelete {
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
		gitopsObj.GetStatus().SetState(status.CLONING_GIT)
	}

	gitopsWatcher.Logger.Info("reconcile")

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

	events.Dispatch(
		events.NewKindEvent(events.EVENT_CHANGED, gitopsWatcher.Gitops.GetDefinition(), nil),
		shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode(),
	)

	if reconcile {
		gitopsWatcher.GitopsQueue <- gitopsObj
	} else {
		switch gitopsObj.GetStatus().GetState() {
		case status.PENDING_DELETE:
			gitopsWatcher.Ticker.Stop()
			gitopsWatcher.Cancel()
			break
		default:
			events.Dispatch(
				events.NewKindEvent(events.EVENT_INSPECT, gitopsWatcher.Gitops.GetDefinition(), nil),
				shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode(),
			)

			if gitopsObj.GetStatus().GetCategory() == status.CATEGORY_END {
				gitopsWatcher.Ticker.Stop()
			} else {
				gitopsWatcher.Ticker.Reset(5 * time.Second)
			}
		}
	}
}
