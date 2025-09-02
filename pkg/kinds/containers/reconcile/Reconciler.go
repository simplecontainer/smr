package reconcile

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/metrics"
	"go.uber.org/zap"
)

func Containers(shared *shared.Shared, containerWatcher *watcher.Container) {
	if containerWatcher.Container.GetStatus().GetPending().Is(status.PENDING_DELETE) {
		return
	}

	containerObj := containerWatcher.Container

	cs := GetState(containerWatcher)

	if !containerObj.GetStatus().IsQueueEmpty() {
		err := containerObj.GetStatus().TransitionToNext()

		if err != nil {
			containerWatcher.Logger.Error("failed to transition state", zap.Error(err))
			return
		}
	}

	err := shared.Registry.Sync(containerObj.GetGroup(), containerObj.GetGeneratedName())
	if err != nil {
		containerWatcher.Logger.Error(err.Error())
	}

	existing := shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
	nextState, reconcile := Reconcile(shared, containerWatcher, existing, cs.State, cs.Error)

	if nextState != "" {
		err = containerObj.GetStatus().QueueState(nextState)
		if err != nil {
			containerWatcher.Logger.Error("failed to queue state", zap.String("state", nextState), zap.Error(err))
		}
	}

	// Do not touch container on this node since it is active on another node
	if containerObj.GetStatus().State.State != status.TRANSFERING {
		cs = GetState(containerWatcher)

		events.Dispatch(
			events.NewKindEvent(containerObj.GetStatus().State.State, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
			shared, containerWatcher.Container.GetRuntime().Node.NodeID,
		)

		events.Dispatch(
			events.NewKindEvent(events.EVENT_CHANGED, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
			shared, containerWatcher.Container.GetRuntime().Node.NodeID,
		)

		if reconcile {
			containerWatcher.ContainerQueue <- containerObj

			go func() {
				metrics.Containers.Get().DeletePartialMatch(prometheus.Labels{"container": containerObj.GetGeneratedName()})
				metrics.Containers.Set(1, containerObj.GetGeneratedName(), nextState)
				metrics.ContainersHistory.Set(1, containerObj.GetGeneratedName(), nextState)
			}()
		} else {
			if containerObj.GetStatus().GetCategory() == status.CATEGORY_END {
				events.Dispatch(
					events.NewKindEvent(events.EVENT_INSPECT, containerWatcher.Container.GetDefinition(), nil),
					shared, containerWatcher.Container.GetDefinition().GetRuntime().GetNode(),
				)
			}

			// Update set for the end since no new call to this function will occur except for delete that is graveyard
			if nextState != "" {
				err = shared.Registry.Sync(containerObj.GetGroup(), containerObj.GetGeneratedName())
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}
			}
		}
	}
}

func GetState(containerWatcher *watcher.Container) state.State {
	engine, err := containerWatcher.Container.GetState()

	if err != nil {
		return state.State{
			Error: err.Error(),
			State: "",
		}
	}

	return engine
}
