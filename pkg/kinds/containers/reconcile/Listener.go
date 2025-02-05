package reconcile

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"time"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container) {
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

			return
		case <-containerWatcher.ContainerQueue:
			containerWatcher.Ticker.Reset(5 * time.Second)
			go Containers(shared, containerWatcher)
			break
		case <-containerWatcher.Ticker.C:
			go Containers(shared, containerWatcher)
			break
		}
	}
}
