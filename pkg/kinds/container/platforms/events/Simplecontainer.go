package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
)

func NewEventsListener(shared *shared.Shared, e chan distributed.KV) {
	for {
		select {
		case data := <-e:
			var event Events

			err := json.Unmarshal(data.Val, &event)

			if err != nil {
				logger.Log.Debug("failed to parse event for processing", zap.String("event", string(data.Val)))
			}

			Event(shared, event)
		}
	}
}

func Event(shared *shared.Shared, event Events) {
	switch event.Kind {
	case EVENT_CHANGE:
		go HandleChange(shared, event)
		break
	case EVENT_RESTART:
		go HandleRestart(shared, event)
		break
	case EVENT_DELETE:
		go HandleDelete(shared, event)
		break
	default:
		break
	}
}

func HandleRestart(shared *shared.Shared, event Events) {
	container := shared.Registry.FindLocal(event.Group, event.Name)

	if container == nil {
		logger.Log.Info("container is not found on this node hence event is ignored for this node")
		return
	}

	if !reconcileIgnore(container.GetLabels()) {
		container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), status.STATUS_CREATED)
		shared.Watcher.Find(container.GetGroupIdentifier()).ContainerQueue <- container
	}
}

func HandleDelete(shared *shared.Shared, event Events) {
	container := shared.Registry.FindLocal(event.Group, event.Name)

	if container == nil {
		logger.Log.Debug("container is not found on this node hence event is ignored for this node")
		return
	}

	if !reconcileIgnore(container.GetLabels()) {
		container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), status.STATUS_PENDING_DELETE)
		shared.Watcher.Find(container.GetGroupIdentifier()).ContainerQueue <- container
	}
}

func HandleChange(shared *shared.Shared, event Events) {
	for _, containerObj := range shared.Watcher.Container {
		if containerObj.Container.HasDependencyOn(event.Kind, event.Group, event.Name, containerObj.Container.GetRuntime()) {
			containerObj.Container.GetStatus().TransitionState(containerObj.Container.GetGroup(), containerObj.Container.GetGeneratedName(), status.STATUS_PENDING_DELETE)
			shared.Watcher.Find(containerObj.Container.GetGroupIdentifier()).ContainerQueue <- containerObj.Container
		}
	}
}
