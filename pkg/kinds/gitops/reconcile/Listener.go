package reconcile

import (
	"github.com/simplecontainer/smr/pkg/events/events"
	f "github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/queue"
	"os"
	"sync"
	"time"
)

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops, pauseHandler func(gitops *watcher.Gitops) error) {
	workerQueue := queue.NewPriorityWorkerQueue(1)
	workerQueue.Start()
	defer workerQueue.Stop()

	lock := &sync.Mutex{}

	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			workerQueue.Submit(queue.WorkTypeCleanup, queue.PriorityCleanup, func() {
				lock.Lock()
				defer lock.Unlock()

				gitopsWatcher.Ticker.Stop()
				gitopsWatcher.Done = true
				close(gitopsWatcher.GitopsQueue)

				var wgChild sync.WaitGroup
				for _, request := range gitopsWatcher.Gitops.Gitops.Pack.Definitions {
					if !request.Definition.Definition.GetState().Gitops.LastSync.IsZero() {
						wgChild.Add(1)

						go func(req *packer.Definition) { // capture request in closure
							defer wgChild.Done()

							format := f.New(
								req.Definition.Definition.GetPrefix(),
								req.Definition.Definition.GetKind(),
								req.Definition.Definition.GetMeta().Group,
								req.Definition.Definition.GetMeta().Name,
							)

							shared.Manager.Replication.Informer.AddCh(format.ToString())

							err := req.Definition.ProposeRemove(
								shared.Manager.Http.Clients[shared.Manager.User.Username].Http,
								shared.Manager.Http.Clients[shared.Manager.User.Username].API,
							)

							if err != nil {
								logger.Log.Error(err.Error())
							}

							select {
							case <-shared.Manager.Replication.Informer.GetCh(format.ToString()):
								return
							}
						}(request)
					}
				}
				wgChild.Wait()

				err := shared.Registry.Remove(
					gitopsWatcher.Gitops.GetDefinition().GetPrefix(),
					gitopsWatcher.Gitops.GetDefinition().GetMeta().Group,
					gitopsWatcher.Gitops.GetDefinition().GetMeta().Name,
				)
				if err != nil {
					logger.Log.Error(err.Error())
				}

				directory := gitopsWatcher.Gitops.Gitops.Git.Directory
				shared.Watchers.Remove(gitopsWatcher.Gitops.GetGroupIdentifier())

				err = os.RemoveAll(directory)
				if err != nil {
					logger.Log.Error(err.Error())
				}

				events.DispatchGroup([]events.Event{
					events.NewKindEvent(events.EVENT_DELETED, gitopsWatcher.Gitops.GetDefinition(), nil),
					events.NewKindEvent(events.EVENT_INSPECT, gitopsWatcher.Gitops.GetDefinition(), nil),
				}, shared, gitopsWatcher.Gitops.GetDefinition().GetRuntime().GetNode())

				gitopsWatcher = nil
				lock = nil
			})

			return

		case <-gitopsWatcher.GitopsQueue:
			workerQueue.Submit(queue.WorkTypeNormal, queue.PriorityNormal, func() {
				if lock == nil || gitopsWatcher == nil {
					return
				}

				lock.Lock()
				defer lock.Unlock()
				Gitops(shared, gitopsWatcher)
			})

		case <-gitopsWatcher.Ticker.C:
			workerQueue.Submit(queue.WorkTypeTicker, queue.PriorityTicker, func() {
				if lock == nil || gitopsWatcher == nil {
					return
				}

				gitopsWatcher.Ticker.Stop()
				lock.Lock()
				defer lock.Unlock()

				if !gitopsWatcher.Done {
					Gitops(shared, gitopsWatcher)
				}
			})

		case <-gitopsWatcher.Poller.C:
			workerQueue.Submit(queue.WorkTypeTicker, queue.PriorityTicker, func() {
				if lock == nil || gitopsWatcher == nil {
					return
				}

				gitopsWatcher.Gitops.SetForceClone(true)
				gitopsWatcher.Ticker.Reset(5 * time.Second)
			})
		}
	}
}
