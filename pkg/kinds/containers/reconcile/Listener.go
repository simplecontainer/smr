package reconcile

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"sync"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
	wg := &sync.WaitGroup{}

	for {
		select {
		case <-containerWatcher.Ctx.Done():
			containerWatcher.Done = true
			containerWatcher.Ticker.Stop()

			close(containerWatcher.ContainerQueue)
			close(containerWatcher.ReadinessChan)
			close(containerWatcher.DependencyChan)

			shared.Registry.Remove(containerWatcher.Container.GetDefinition().GetPrefix(), containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName())
			shared.Watchers.Remove(containerWatcher.Container.GetGroupIdentifier())

			DispatchEventDelete(shared, containerWatcher.Container)
			DispatchEventInspect(shared, containerWatcher.Container)

			containerWatcher = nil
			return
		case <-containerWatcher.ContainerQueue:
			wg.Wait()
			go Containers(shared, containerWatcher, wg)
			break
		case <-containerWatcher.Ticker.C:
			containerWatcher.Ticker.Stop()
			wg.Wait()
			go Containers(shared, containerWatcher, wg)
			break
		case <-containerWatcher.PauseC:
			if pauseHandler(containerWatcher) != nil {
				containerWatcher.Cancel()
			}
			break
		}
	}
}
