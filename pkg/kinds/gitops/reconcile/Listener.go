package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"sync"
	"time"
)

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops, pauseHandler func(gitops *watcher.Gitops) error) {
	wg := &sync.WaitGroup{}

	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			gitopsWatcher.Done = true
			close(gitopsWatcher.GitopsQueue)

			var wgChild sync.WaitGroup
			for _, request := range gitopsWatcher.Gitops.Definitions {
				if !request.Definition.GetState().Gitops.LastSync.IsZero() {
					wgChild.Add(1)

					go func() {
						format := f.New(request.Definition.GetPrefix(), request.Definition.GetKind(), request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
						shared.Manager.Replication.NewDeleteC(format)

						request.Definition.Definition.GetRuntime().SetOwner(gitopsWatcher.Gitops.Definition.GetKind(), gitopsWatcher.Gitops.Definition.GetMeta().Group, gitopsWatcher.Gitops.Definition.GetMeta().Name)
						err := request.ProposeRemove(shared.Manager.Http.Clients[shared.Manager.User.Username].Http, shared.Manager.Http.Clients[shared.Manager.User.Username].API)

						if err != nil {
							logger.Log.Error(err.Error())
						}

						for {
							select {
							case <-shared.Manager.Replication.DeleteC[format.ToString()]:
								close(shared.Manager.Replication.DeleteC[format.ToString()])
								delete(shared.Manager.Replication.DeleteC, format.ToString())

								wgChild.Done()
								break
							}
						}
					}()
				}
			}
			wgChild.Wait()

			err := shared.Registry.Remove(gitopsWatcher.Gitops.Definition.GetPrefix(), gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name)
			if err != nil {
				logger.Log.Error(err.Error())
			}

			shared.Watcher.Remove(fmt.Sprintf("%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))

			events.DispatchGroup([]events.Event{
				events.NewKindEvent(events.EVENT_DELETED, gitopsWatcher.Gitops.GetDefinition(), nil),
				events.NewKindEvent(events.EVENT_INSPECT, gitopsWatcher.Gitops.GetDefinition(), nil),
			}, shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode())

			gitopsWatcher = nil
			return
		case <-gitopsWatcher.GitopsQueue:
			wg.Wait()
			if !gitopsWatcher.Done {
				go Gitops(shared, gitopsWatcher, wg)
			}
			break
		case <-gitopsWatcher.Ticker.C:
			gitopsWatcher.Ticker.Stop()
			wg.Wait()
			if !gitopsWatcher.Done {
				go Gitops(shared, gitopsWatcher, wg)
			}
			break
		case <-gitopsWatcher.Poller.C:
			gitopsWatcher.Gitops.ForcePoll = true
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			break
		}
	}
}
