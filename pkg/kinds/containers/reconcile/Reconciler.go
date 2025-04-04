package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"go.uber.org/zap"
)

func Containers(shared *shared.Shared, containerWatcher *watcher.Container) {
	if containerWatcher.Container.GetStatus().GetPending().Is(status.PENDING_DELETE) {
		return
	}

	containerObj := containerWatcher.Container

	cs := GetState(containerWatcher)

	existing := shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
	newState, reconcile := Reconcile(shared, containerWatcher, existing, cs.State, cs.Error)

	// Do not touch container on this node since it is active on another node
	if containerObj.GetStatus().State.State != status.TRANSFERING {
		if containerObj.GetStatus().State.State != newState {
			transitioned := containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetName(), newState)

			if !transitioned {
				containerWatcher.Logger.Error("failed to transition state",
					zap.String("old", containerObj.GetStatus().State.State),
					zap.String("new", newState))
			}
		}

		cs = GetState(containerWatcher)

		err := shared.Registry.Sync(containerObj.GetGroup(), containerObj.GetGeneratedName())

		if err != nil {
			containerWatcher.Logger.Error(err.Error())
		}

		events.Dispatch(
			events.NewKindEvent(events.EVENT_CHANGED, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
			shared, containerWatcher.Container.GetRuntime().Node.NodeID,
		)

		if newState == "" {
			containerWatcher.Container.GetStatus().GetPending().Set(status.PENDING_DELETE)
			containerWatcher.Cancel()

			return
		}

		if reconcile {
			containerWatcher.ContainerQueue <- containerObj
		} else {
			switch containerObj.GetStatus().GetState() {
			default:
				if containerObj.GetStatus().GetCategory() == status.CATEGORY_END {
					events.Dispatch(
						events.NewKindEvent(events.EVENT_INSPECT, containerWatcher.Container.GetDefinition(), nil),
						shared, containerWatcher.Container.GetDefinition().GetRuntime().GetNode(),
					)
				}

				containerWatcher.Ticker.Stop()
			}
		}
	}
}

func GetState(containerWatcher *watcher.Container) state.State {
	engine, err := containerWatcher.Container.GetState()

	if err != nil {
		return state.State{}
	}

	return engine
}
