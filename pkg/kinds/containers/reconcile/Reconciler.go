package reconcile

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"go.uber.org/zap"
)

func Containers(shared *shared.Shared, containerWatcher *watcher.Container) {
	if containerWatcher.Done {
		return
	}

	containerObj := containerWatcher.Container

	if containerObj.GetStatus().Reconciling {
		containerWatcher.Logger.Info("container already reconciling, waiting for the free slot")
		return
	}

	containerObj.GetStatus().Reconciling = true

	state := GetState(containerWatcher)

	if containerObj.GetStatus().State.State == status.STATUS_DAEMON_FAILURE {
		shared.Registry.Sync(containerObj)

		containerWatcher.Logger.Info("reconciler is going to sleep - runtime daemon error")
		containerWatcher.Ticker.Stop()
	} else {
		existing := shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
		newState, reconcile := Reconcile(shared, containerWatcher, existing, state.State, state.Error)

		containerObj.GetStatus().Reconciling = false

		if containerObj.GetStatus().State.State != newState {
			transitioned := containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetName(), newState)

			if !transitioned {
				containerWatcher.Logger.Error("failed to transition state",
					zap.String("old", containerObj.GetStatus().State.State),
					zap.String("new", newState))
			}
		}

		shared.Registry.Sync(containerObj)

		if containerObj.GetStatus().GetCategory() == status.CATEGORY_END {
			containerWatcher.Ticker.Stop()
		}

		if reconcile {
			Containers(shared, containerWatcher)
		}
	}
}

func GetState(containerWatcher *watcher.Container) state.State {
	containerStateEngine, err := containerWatcher.Container.GetContainerState()

	if err != nil {
		return state.State{}
	}

	return containerStateEngine
}
