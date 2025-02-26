package reconcile

import (
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

			DispatchEventDelete(shared, containerWatcher.Container, containerWatcher.Container.GetGeneratedName())

			replicas := make([]platforms.IContainer, 0)
			group := shared.Registry.FindGroup(containerWatcher.Container.GetDefinition().GetPrefix(), containerWatcher.Container.GetGroup())

			for _, c := range group {
				if c.GetName() == containerWatcher.Container.GetName() {
					replicas = append(replicas, c)
				}
			}

			if len(replicas) == 0 {
				DispatchEventDelete(shared, containerWatcher.Container, containerWatcher.Container.GetName())
				DispatchEventInspect(shared, containerWatcher.Container)
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
