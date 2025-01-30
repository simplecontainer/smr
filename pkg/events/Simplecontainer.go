package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	containerShared "github.com/simplecontainer/smr/pkg/kinds/container/shared"
	containerStatus "github.com/simplecontainer/smr/pkg/kinds/container/status"
	gitopsShared "github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	gitopsStatus "github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
)

func NewEventsListener(kr map[string]contracts.Kind, e chan KV.KV) {
	for {
		select {
		case data := <-e:
			var event Events

			format := f.NewUnformated(data.Key, static.CATEGORY_DNS_STRING)
			acks.ACKS.Ack(format.GetUUID())

			err := json.Unmarshal(data.Val, &event)

			if err != nil {
				logger.Log.Debug("failed to parse event for processing", zap.String("event", string(data.Val)))
			}

			Event(kr, event, data.Node)
		}
	}
}

func Event(kr map[string]contracts.Kind, event Events, node uint64) {
	switch event.Type {
	case EVENT_CHANGE:
		go HandleChange(kr["container"].GetShared().(*containerShared.Shared), event, node)
		break
	case EVENT_RESTART:
		go HandleRestart(kr["container"].GetShared().(*containerShared.Shared), event, node)
		break
	case EVENT_DELETE:
		go HandleDelete(kr["container"].GetShared().(*containerShared.Shared), event, node)
		break
	case EVENT_SYNC:
		go HandleSync(kr["gitops"].GetShared().(*gitopsShared.Shared), event, node)
		break
	case EVENT_REFRESH:
		go HandleRefresh(kr["gitops"].GetShared().(*gitopsShared.Shared), event, node)
		break
	default:
		break
	}
}

func HandleRestart(shared *containerShared.Shared, event Events, node uint64) {
	container := shared.Registry.FindLocal(event.Group, event.Name)

	if container == nil {
		logger.Log.Info("container is not found on this node hence event is ignored for this node")
		return
	}

	if !reconcileIgnore(container.GetLabels()) {
		container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), containerStatus.STATUS_CREATED)
		shared.Watcher.Find(container.GetGroupIdentifier()).ContainerQueue <- container
	}
}

func HandleDelete(shared *containerShared.Shared, event Events, node uint64) {
	container := shared.Registry.FindLocal(event.Group, event.Name)

	if container == nil {
		logger.Log.Debug("container is not found on this node hence event is ignored for this node")
		return
	}

	if !reconcileIgnore(container.GetLabels()) {
		container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), containerStatus.STATUS_PENDING_DELETE)
		shared.Watcher.Find(container.GetGroupIdentifier()).ContainerQueue <- container
	}
}

func HandleChange(shared *containerShared.Shared, event Events, node uint64) {
	for _, containerWatcher := range shared.Watcher.Container {
		if containerWatcher.Container.HasDependencyOn(event.Kind, event.Group, event.Name) {
			containerWatcher.Container.GetStatus().TransitionState(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName(), containerStatus.STATUS_CHANGE)
			shared.Watcher.Find(containerWatcher.Container.GetGroupIdentifier()).ContainerQueue <- containerWatcher.Container
		}
	}
}

func HandleRefresh(shared *gitopsShared.Shared, event Events, node uint64) {
	gitops := shared.Registry.FindLocal(event.Group, event.Name)

	if gitops == nil {
		logger.Log.Debug("container is not found on this node hence event is ignored for this node")
		return
	}

	gitopsWatcher := shared.Watcher.Find(gitops.GetGroupIdentifier())

	if gitopsWatcher != nil {
		gitopsWatcher.Gitops.ForcePoll = true
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, gitopsStatus.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}
}

func HandleSync(shared *gitopsShared.Shared, event Events, node uint64) {
	gitops := shared.Registry.FindLocal(event.Group, event.Name)

	if gitops == nil {
		logger.Log.Debug("container is not found on this node hence event is ignored for this node")
		return
	}

	gitopsWatcher := shared.Watcher.Find(gitops.GetGroupIdentifier())

	if gitopsWatcher != nil {
		gitopsWatcher.Gitops.ManualSync = true

		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, gitopsStatus.STATUS_CLONING_GIT)
		gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
	}
}
