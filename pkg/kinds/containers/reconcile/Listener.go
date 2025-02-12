package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
)

func HandleTickerAndEvents(shared *shared.Shared, containerWatcher *watcher.Container, pauseHandler func(*watcher.Container) error) {
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
			containerWatcher = nil

			return
		case <-containerWatcher.ContainerQueue:
			fmt.Println("listener got container - firing reconcile")
			go Containers(shared, containerWatcher)
			break
		case <-containerWatcher.Ticker.C:
			fmt.Println("ticker firing reconcile")
			go Containers(shared, containerWatcher)
			break
		case <-containerWatcher.PauseC:
			if pauseHandler(containerWatcher) != nil {
				containerWatcher.Cancel()
			}
			break
		}
	}
}
