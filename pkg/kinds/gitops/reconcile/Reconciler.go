package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"go.uber.org/zap"
	"time"
)

func Gitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	if gitopsWatcher.Done || gitopsWatcher.Gitops.Gitops.Status.PendingDelete {
		return
	}

	gitopsObj := gitopsWatcher.Gitops

	if gitopsObj.Gitops.ForceClone {
		gitopsObj.SetForceClone(false)
		gitopsObj.GetStatus().SetState(status.CLONING_GIT)
	}

	gitopsWatcher.Logger.Info("reconcile", zap.String("gitops", gitopsObj.GetName()), zap.String("status", fmt.Sprintf("%v", gitopsObj.GetStatus().State)))

	newState, reconcile := Reconcile(shared, gitopsWatcher)

	transitioned := gitopsObj.GetStatus().TransitionState(gitopsObj.GetGroup(), gitopsObj.GetName(), newState)

	if !transitioned && newState != "" {
		gitopsWatcher.Logger.Error("failed to transition state",
			zap.String("old", gitopsObj.GetStatus().State.State),
			zap.String("new", newState))
	}

	err := shared.Registry.Sync(gitopsObj.GetGroup(), gitopsObj.GetName())

	if err != nil {
		gitopsWatcher.Logger.Error(err.Error())
	}

	events.Dispatch(
		events.NewKindEvent(gitopsWatcher.Gitops.GetStatus().GetState(), gitopsWatcher.Gitops.GetDefinition(), nil),
		shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode(),
	)

	events.Dispatch(
		events.NewKindEvent(events.EVENT_CHANGED, gitopsWatcher.Gitops.GetDefinition(), nil),
		shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode(),
	)

	if reconcile {
		gitopsWatcher.GitopsQueue <- gitopsObj
	} else {
		switch gitopsObj.GetStatus().GetState() {
		case status.DRIFTED, status.INSYNC:
			gitopsWatcher.Gitops.Gitops.Status.GetPending().Clear()
			gitopsWatcher.Ticker.Stop()
		case status.DELETE:
			gitopsWatcher.Ticker.Stop()
			gitopsWatcher.Cancel()
			break
		default:
			if gitopsObj.GetStatus().GetCategory() == status.CATEGORY_END {
				gitopsWatcher.Gitops.Gitops.Status.GetPending().Clear()

				events.Dispatch(
					events.NewKindEvent(events.EVENT_INSPECT, gitopsWatcher.Gitops.GetDefinition(), nil),
					shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode(),
				)

				gitopsWatcher.Ticker.Stop()
			} else {
				gitopsWatcher.Ticker.Reset(5 * time.Second)
			}
		}
	}
}
