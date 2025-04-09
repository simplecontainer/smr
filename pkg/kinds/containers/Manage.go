package containers

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
)

func (containers *Containers) Create(cs []platforms.IContainer, exists bool, user *authentication.User) {
	for _, containerObj := range cs {
		if exists {
			existingWatcher := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

			if existingWatcher != nil {
				// forbiden create occured
			} else {
				// container object holding multiple replicas existed but watcher was never born
				// this means that replica landed on this node first time, could be that it comes from 2nd node
				// implication:
				// - create watcher
				// - assing container to the wathcer
				// - roll the reconciler

				existingContainer := containers.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

				if existingContainer != nil && existingContainer.IsGhost() {
					w := watcher.New(containerObj, status.TRANSFERING, user)
					containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
					containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

					w.Logger.Info("container object created")

					go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
						return nil
					})

					go reconcile.Containers(containers.Shared, w)
				} else {
					w := watcher.New(containerObj, status.CREATED, user)
					containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
					containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

					w.Logger.Info("container object created")

					go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
						return nil
					})
					go reconcile.Containers(containers.Shared, w)
				}
			}
		} else {
			// container object holding multiple replicas never existed
			// implication:
			// - create watcher
			// - assing container to the wathcer
			// - roll the reconciler

			w := watcher.New(containerObj, status.CREATED, user)
			containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
			containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

			w.Logger.Info("container object created")

			go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
				return nil
			})

			go reconcile.Containers(containers.Shared, w)
		}
	}
}

func (containers *Containers) Update(cs []platforms.IContainer, exists bool) {
	for _, containerObj := range cs {
		if exists {
			existingWatcher := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

			if existingWatcher != nil {
				existingWatcher.Logger.Info("container object modified, reusing watcher")
				existingContainer := containers.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

				if existingContainer != nil {
					existingWatcher.Ticker.Stop()

					containerObj.GetStatus().SetState(status.CREATED)
					containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

					existingWatcher.Container = containerObj
					go reconcile.Containers(containers.Shared, existingWatcher)
				}
			} else {
				existingWatcher.Logger.Info("no changes detected on the container object")
			}
		} else {
			// forbiden update occured
		}
	}
}

func (containers *Containers) Destroy(cs []platforms.IContainer, exists bool) {
	for _, containerObj := range cs {
		GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
		containerW := containers.Shared.Watchers.Find(GroupIdentifier)

		if containerW != nil && !containerW.Done {
			containers.Shared.Watchers.Find(GroupIdentifier).DeleteC <- containerObj
		}
	}
}
