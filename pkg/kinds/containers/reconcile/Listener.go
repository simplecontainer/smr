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
	"time"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
	workerQueue := queue.NewPriorityWorkerQueue(1)
	workerQueue.Start()

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("panic in HandleTickerAndEvents", zap.Any("panic", r))
		}
		workerQueue.Stop()
	}()

	lock := &sync.Mutex{}
	cleanupOnce := &sync.Once{}

	cleanup := func() {
		cleanupOnce.Do(func() {
			lock.Lock()
			defer lock.Unlock()

			if containerWatcher == nil {
				return
			}

			containerWatcher.Logger.Info("cleaning up container")
			containerWatcher.SetDone(true)
			containerWatcher.SafeStopTicker()

			safeCloseContainer := func() {
				defer func() {
					if r := recover(); r != nil {
						containerWatcher.Logger.Debug("channel already closed (ContainerQueue)")
					}
				}()
				close(containerWatcher.ContainerQueue)
			}

			safeCloseReadiness := func() {
				defer func() {
					if r := recover(); r != nil {
						containerWatcher.Logger.Debug("channel already closed (ReadinessChan)")
					}
				}()
				close(containerWatcher.ReadinessChan)
			}

			safeCloseDependency := func() {
				defer func() {
					if r := recover(); r != nil {
						containerWatcher.Logger.Debug("channel already closed (DependencyChan)")
					}
				}()
				close(containerWatcher.DependencyChan)
			}

			safeCloseContainer()
			safeCloseReadiness()
			safeCloseDependency()

			err := containerWatcher.Container.Clean()
			if err != nil {
				containerWatcher.Logger.Error(err.Error())
			}

			containerWatcher.Logger.Info("container cleaned up")

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
		})
	}

	for {
		select {
		case <-containerWatcher.Ctx.Done():
			workerQueue.Submit(queue.WorkTypeCleanup, queue.PriorityCleanup, func() {
				cleanup()
			})
			return

		case <-containerWatcher.ContainerQueue:
			cw := containerWatcher
			currentLock := lock

			if cw.IsDone() {
				continue
			}

			workerQueue.Submit(queue.WorkTypeNormal, queue.PriorityNormal, func() {
				if currentLock == nil || cw == nil || cw.IsDone() {
					return
				}

				cw.SafeStopTicker()

				currentLock.Lock()
				defer currentLock.Unlock()

				if cw.IsDone() {
					return
				}

				Containers(shared, cw)
			})

		case <-containerWatcher.Ticker.C:
			// ✅ Capture variables
			cw := containerWatcher
			currentLock := lock

			if cw.IsDone() {
				continue
			}

			workerQueue.Submit(queue.WorkTypeTicker, queue.PriorityTicker, func() {
				if currentLock == nil || cw == nil || cw.IsDone() {
					return
				}

				cw.SafeStopTicker()

				currentLock.Lock()
				defer currentLock.Unlock()

				if cw.IsDone() {
					return
				}

				Containers(shared, cw)
			})

		case <-containerWatcher.DeleteC:
			// ✅ Capture variables
			cw := containerWatcher
			currentLock := lock

			if cw.IsDone() {
				continue
			}

			if pauseHandler(cw) == nil {
				cw.SetAllowPlatformEvents(false)
				cw.ChecksCancel()
				cw.ReconcileCancel()

				workerQueue.Submit(queue.WorkTypeDelete, queue.PriorityDelete, func() {
					if currentLock == nil || cw == nil {
						return
					}

					currentLock.Lock()
					defer currentLock.Unlock()

					if cw.IsDone() {
						return
					}

					cw.Container.GetStatus().RejectQueueAttempts(time.Now())
					cw.Container.GetStatus().ClearQueue()
					cw.Container.GetStatus().QueueState(status.DELETE, time.Now())
					Containers(shared, cw)
				})
			}
		}
	}
}
