package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/implementations/containers/shared"
	"github.com/simplecontainer/smr/implementations/containers/watcher"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/plugins"
	"time"
)

func NewWatcher(containers v1.Containers) *watcher.Containers {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	return &watcher.Containers{
		Definition:      containers,
		Syncing:         false,
		Tracking:        false,
		ContainersQueue: make(chan string),
		Ctx:             ctx,
		Cancel:          fn,
		Ticker:          time.NewTicker(interval),
	}
}

func HandleTickerAndEvents(shared *shared.Shared, containers *watcher.Containers) {
	for {
		select {
		case <-containers.Ctx.Done():
			containers.Ticker.Stop()
			close(containers.ContainersQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", containers.Definition.Meta.Group, containers.Definition.Meta.Name))

			return
		case _ = <-containers.ContainersQueue:
			ReconcileContainer(shared, containers)
			break
		case _ = <-containers.Ticker.C:
			if !containers.Syncing {
				ReconcileContainer(shared, containers)
			}
			break
		}
	}
}

func ReconcileContainer(shared *shared.Shared, containers *watcher.Containers) {
	if containers.Syncing {
		logger.Log.Info("containers already reconciling, waiting for the free slot")
		return
	}

	containers.Syncing = true

	for _, definition := range containers.Definition.Spec {
		definitionString, _ := definition.ToJsonString()

		pl := plugins.GetPlugin(shared.Manager.Config.Root, "container.so")
		pl.Apply([]byte(definitionString))
	}

	containers.Syncing = false
}
