package containers

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"time"
)

func (containers *Containers) Create(cs []platforms.IContainer, exists bool, user *authentication.User) {
	for _, containerObj := range cs {
		groupIdentifier := containerObj.GetGroupIdentifier()
		existingWatcher := containers.Shared.Watchers.Find(groupIdentifier)

		if !exists || existingWatcher == nil {
			s := status.CREATED

			if exists && containerObj.IsGhost() {
				s = status.TRANSFERING
			}

			w := watcher.New(containerObj, s, user)
			containers.Shared.Watchers.AddOrUpdate(groupIdentifier, w)
			containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

			containerObj.GetStatus().QueueState(status.CREATED, time.Now())
			w.Logger.Info("container object created")

			go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
				return nil
			})
			go reconcile.Containers(containers.Shared, w)
		} else {
			existingWatcher.Logger.Info("container already exists, forbidden create")
		}
	}
}

func (containers *Containers) Update(cs []platforms.IContainer, exists bool) {
	for _, containerObj := range cs {
		groupIdentifier := containerObj.GetGroupIdentifier()
		existingWatcher := containers.Shared.Watchers.Find(groupIdentifier)

		if existingWatcher != nil {
			existingWatcher.Logger.Info("container object modified, reusing watcher")

			existingContainer := containers.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
			if existingContainer != nil {
				existingWatcher.Ticker.Stop()

				containerObj.GetStatus().QueueState(status.CREATED, time.Now())
				containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

				existingWatcher.Container = containerObj
				go reconcile.Containers(containers.Shared, existingWatcher)
			}
		} else {
			existingWatcher.Logger.Info("no changes detected on the container object")
		}
	}
}

func (containers *Containers) Destroy(cs []platforms.IContainer, exists bool) {
	for _, containerObj := range cs {
		containerW := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

		if containerW != nil && !containerW.IsDone() {
			sent := containerW.SendDelete(containerObj, 5*time.Second)
			if !sent {
				containerW.Logger.Warn("failed to send delete signal")
			}
		}
	}
}
