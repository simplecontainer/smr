package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"sync"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
	lock := &sync.Mutex{}

	for {
		select {
		case <-containerWatcher.Ctx.Done():
			lock.Lock()
			containerWatcher.Ticker.Stop()

			containerWatcher.Done = true

			close(containerWatcher.ContainerQueue)
			close(containerWatcher.ReadinessChan)
			close(containerWatcher.DependencyChan)

			err := containerWatcher.Container.Clean()

			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}

			err = shared.Registry.Remove(containerWatcher.Container.GetDefinition().GetPrefix(), containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName())

			if err != nil {
				logger.Log.Error("failed to remove container state", zap.Error(err))
			}

			shared.Watchers.Remove(containerWatcher.Container.GetGroupIdentifier())

			events.Dispatch(
				events.NewKindEvent(events.EVENT_DELETED, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
				shared, containerWatcher.Container.GetRuntime().Node.NodeID,
			)

			replicas := make([]platforms.IContainer, 0)
			group := shared.Registry.FindGroup(containerWatcher.Container.GetDefinition().GetPrefix(), containerWatcher.Container.GetGroup())

			for _, c := range group {
				if c.GetName() == containerWatcher.Container.GetName() {
					replicas = append(replicas, c)
				}
			}

			if len(replicas) == 0 {
				events.DispatchGroup([]events.Event{
					events.NewKindEvent(events.EVENT_DELETED, containerWatcher.Container.GetDefinition(), nil),
					events.NewKindEvent(events.EVENT_INSPECT, containerWatcher.Container.GetDefinition(), nil),
					events.NewKindEvent(events.EVENT_DEPENDENCY, containerWatcher.Container.GetDefinition(), nil),
				}, shared, containerWatcher.Container.GetDefinition().GetRuntime().GetNode())
			}

			containerWatcher = nil
			lock = nil
			return
		case <-containerWatcher.ContainerQueue:
			go func() {
				lock.Lock()
				Containers(shared, containerWatcher)
				lock.Unlock()
			}()
			break
		case <-containerWatcher.Ticker.C:
			go func() {
				containerWatcher.Ticker.Stop()
				lock.Lock()

				if !containerWatcher.Done {
					Containers(shared, containerWatcher)
				}

				lock.Unlock()
			}()
			break
		case <-containerWatcher.DeleteC:
			if pauseHandler(containerWatcher) == nil {
				go func() {
					containerWatcher.ReconcileCancel()

					lock.Lock()

					containerWatcher.Container.GetStatus().TransitionState(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName(), status.DELETE)

					if !containerWatcher.Done {
						Containers(shared, containerWatcher)
					}

					lock.Unlock()
				}()
			}
			break
		}
	}
}
