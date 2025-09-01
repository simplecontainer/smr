package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/queue"
	"go.uber.org/zap"
	"sync"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
	workerQueue := queue.NewPriorityWorkerQueue(1)
	workerQueue.Start()
	defer workerQueue.Stop()

	lock := &sync.Mutex{}

	for {
		select {
		case <-containerWatcher.Ctx.Done():
			workerQueue.Submit(queue.WorkTypeCleanup, queue.PriorityCleanup, func() {
				lock.Lock()
				defer lock.Unlock()

				containerWatcher.Logger.Info("cleaning up container")

				containerWatcher.Ticker.Stop()
				containerWatcher.Done = true

				close(containerWatcher.ContainerQueue)
				close(containerWatcher.ReadinessChan)
				close(containerWatcher.DependencyChan)

				err := containerWatcher.Container.Clean()
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
				}

				err = shared.Registry.Remove(
					containerWatcher.Container.GetDefinition().GetPrefix(),
					containerWatcher.Container.GetGroup(),
					containerWatcher.Container.GetGeneratedName(),
				)
				if err != nil {
					logger.Log.Error("failed to remove container state", zap.Error(err))
				}

				shared.Watchers.Remove(containerWatcher.Container.GetGroupIdentifier())

				events.DispatchGroup([]events.Event{
					events.NewKindEvent(events.EVENT_DELETED, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
					events.NewKindEvent(events.EVENT_CHANGE, containerWatcher.Container.GetDefinition(), nil).SetName(containerWatcher.Container.GetGeneratedName()),
				}, shared, containerWatcher.Container.GetRuntime().Node.NodeID)

				replicas := make([]platforms.IContainer, 0)
				group := shared.Registry.FindGroup(
					containerWatcher.Container.GetDefinition().GetPrefix(),
					containerWatcher.Container.GetGroup(),
				)

				for _, c := range group {
					if c.GetName() == containerWatcher.Container.GetName() {
						replicas = append(replicas, c)
					}
				}

				if len(replicas) == 0 {
					events.DispatchGroup([]events.Event{
						events.NewKindEvent(events.EVENT_DELETED, containerWatcher.Container.GetDefinition(), nil),
						events.NewKindEvent(events.EVENT_INSPECT, containerWatcher.Container.GetDefinition(), nil),
					}, shared, containerWatcher.Container.GetDefinition().GetRuntime().GetNode())
				}

				containerWatcher.Logger.Info("container cleaned up")
				containerWatcher = nil
			})

			return

		case <-containerWatcher.ContainerQueue:
			if !containerWatcher.Done {
				workerQueue.Submit(queue.WorkTypeNormal, queue.PriorityNormal, func() {
					if lock == nil || containerWatcher == nil {
						return
					}

					lock.Lock()
					defer lock.Unlock()
					Containers(shared, containerWatcher)
				})
			}

		case <-containerWatcher.Ticker.C:
			workerQueue.Submit(queue.WorkTypeTicker, queue.PriorityTicker, func() {
				if lock == nil || containerWatcher == nil {
					return
				}

				containerWatcher.Ticker.Stop()
				lock.Lock()
				defer lock.Unlock()

				if !containerWatcher.Done {
					Containers(shared, containerWatcher)
				}
			})

		case <-containerWatcher.DeleteC:
			if pauseHandler(containerWatcher) == nil {
				containerWatcher.AllowPlatformEvents = false
				containerWatcher.ReconcileCancel()

				workerQueue.Submit(queue.WorkTypeDelete, queue.PriorityDelete, func() {
					if lock == nil || containerWatcher == nil {
						return
					}

					lock.Lock()
					defer lock.Unlock()

					containerWatcher.Container.GetStatus().ClearQueue()
					containerWatcher.Container.GetStatus().QueueState(status.DELETE)
					Containers(shared, containerWatcher)
				})
			}
		}
	}
}
