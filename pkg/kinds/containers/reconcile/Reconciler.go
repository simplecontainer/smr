package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containers v1.ContainersDefinition, mgr *manager.Manager) *watcher.Containers {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{fmt.Sprintf("/tmp/containers.%s.%s.log", containers.Meta.Group, containers.Meta.Name)}

	loggerObj, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return &watcher.Containers{
		Definition:      containers,
		Syncing:         false,
		Tracking:        false,
		ContainersQueue: make(chan string),
		Ctx:             ctx,
		Cancel:          fn,
		Ticker:          time.NewTicker(interval),
		Logger:          loggerObj,
	}
}

func HandleTickerAndEvents(shared *shared.Shared, user *authentication.User, containers *watcher.Containers, agent string) {
	for {
		select {
		case <-containers.Ctx.Done():
			containers.Ticker.Stop()
			close(containers.ContainersQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", containers.Definition.Meta.Group, containers.Definition.Meta.Name))

			return
		case _ = <-containers.ContainersQueue:
			Container(shared, user, containers, agent)
			break
		}
	}
}

func Container(shared *shared.Shared, user *authentication.User, containers *watcher.Containers, agent string) {
	if containers.Syncing {
		containers.Logger.Info("containers already reconciling, waiting for the free slot")
		return
	}

	containers.Syncing = true

	for _, definition := range containers.Definition.Spec {
		definitionJSON, err := definition.ToJson()

		if err != nil {
			containers.Logger.Info(err.Error())
		} else {
			_, err = shared.Manager.KindsRegistry["container"].Apply(user, definitionJSON, agent)

			if err != nil {
				logger.Log.Error(err.Error())
				return
			}
		}
	}

	containers.Syncing = false
}
