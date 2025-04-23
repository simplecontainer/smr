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
	lock := &sync.Mutex{}

	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			gitopsWatcher.Done = true
			close(gitopsWatcher.GitopsQueue)

			var wgChild sync.WaitGroup
			for _, request := range gitopsWatcher.Gitops.Pack.Definitions {
				if !request.Definition.GetState().Gitops.LastSync.IsZero() {
					wgChild.Add(1)

					go func() {
						format := f.New(request.Definition.GetPrefix(), request.Definition.GetKind(), request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
						shared.Manager.Replication.Informer.AddCh(format.ToString())

						err := request.ProposeRemove(shared.Manager.Http.Clients[shared.Manager.User.Username].Http, shared.Manager.Http.Clients[shared.Manager.User.Username].API)

						if err != nil {
							logger.Log.Error(err.Error())
						}

						select {
						case <-shared.Manager.Replication.Informer.GetCh(format.ToString()):
							shared.Manager.Replication.Informer.RmCh(format.ToString())
							wgChild.Done()
							break
						}
					}()
				}
			}
			wgChild.Wait()

			err := shared.Registry.Remove(gitopsWatcher.Gitops.Definition.GetPrefix(), gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name)
			if err != nil {
				logger.Log.Error(err.Error())
			}

			shared.Watchers.Remove(fmt.Sprintf("%s/%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))

			events.DispatchGroup([]events.Event{
				events.NewKindEvent(events.EVENT_DELETED, gitopsWatcher.Gitops.GetDefinition(), nil),
				events.NewKindEvent(events.EVENT_INSPECT, gitopsWatcher.Gitops.GetDefinition(), nil),
			}, shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode())

			gitopsWatcher = nil
			return
		case <-gitopsWatcher.GitopsQueue:
			go func() {
				lock.Lock()
				Gitops(shared, gitopsWatcher)
				lock.Unlock()
			}()
			break
		case <-gitopsWatcher.Ticker.C:
			go func() {
				gitopsWatcher.Ticker.Stop()
				lock.Lock()

				if !gitopsWatcher.Done {
					Gitops(shared, gitopsWatcher)
				}

				lock.Unlock()
			}()
			break
		case <-gitopsWatcher.Poller.C:
			gitopsWatcher.Gitops.ForceClone = true
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			break
		}
	}
}
