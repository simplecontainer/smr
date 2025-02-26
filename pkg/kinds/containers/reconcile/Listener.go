package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"sync"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
	wg := &sync.WaitGroup{}

	for {
		select {
		case <-containerWatcher.Ctx.Done():
			wg.Wait()
			containerWatcher.Ticker.Stop()

			containerWatcher.Done = true

			close(containerWatcher.ContainerQueue)
			close(containerWatcher.ReadinessChan)
			close(containerWatcher.DependencyChan)

			shared.Registry.Remove(containerWatcher.Container.GetDefinition().GetPrefix(), containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName())
			shared.Watchers.Remove(containerWatcher.Container.GetGroupIdentifier())

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
				}, shared, containerWatcher.Container.GetDefinition().GetRuntime().GetNode())
			}

			containerWatcher = nil
			wg = nil
			return
		case <-containerWatcher.ContainerQueue:
			wg.Wait()
			go Containers(shared, containerWatcher, wg)
			break
		case <-containerWatcher.Ticker.C:
			containerWatcher.Ticker.Stop()
			wg.Wait()
			if !containerWatcher.Done {
				go Containers(shared, containerWatcher, wg)
			}
			break
		case <-containerWatcher.PauseC:
			if pauseHandler(containerWatcher) != nil {
				containerWatcher.Cancel()
			}
			break
		}
	}
}
