package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"time"
)

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops, pauseHandler func(gitops *watcher.Gitops) error) {
	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			close(gitopsWatcher.GitopsQueue)

			logger.Log.Debug("gitops watcher deleted - proceed with deleting children")

			for _, request := range gitopsWatcher.Children {
				go func() {
					err := request.ProposeRemove(shared.Manager.Http.Clients[shared.Manager.User.Username].Http, shared.Manager.Http.Clients[shared.Manager.User.Username].API)

					if err != nil {
						logger.Log.Error(err.Error())
					}
				}()
			}

			shared.Registry.Remove(gitopsWatcher.Gitops.Definition.GetPrefix(), gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))

			return
		case <-gitopsWatcher.GitopsQueue:
			go Gitops(shared, gitopsWatcher)
			break
		case <-gitopsWatcher.Ticker.C:
			gitopsWatcher.Ticker.Stop()
			go Gitops(shared, gitopsWatcher)
			break
		case <-gitopsWatcher.Poller.C:
			gitopsWatcher.Gitops.ForcePoll = true
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			break
		}
	}
}
