package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
	"strings"
	"time"
)

func NewWatcher(containers v1.ContainersDefinition, mgr *manager.Manager) *watcher.Containers {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	format := f.New(containers.GetPrefix(), "kind", static.KIND_CONTAINERS, containers.Meta.Group, containers.Meta.Name)
	path := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{path}, []string{path})

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
		definition.SetRuntime(&commonv1.Runtime{})
		definition.GetRuntime().SetOwner(static.KIND_CONTAINERS, containers.Definition.Meta.Group, containers.Definition.Meta.Name)
		definition.GetRuntime().SetNode(containers.Definition.GetRuntime().GetNode())
		definition.Kind = static.KIND_CONTAINER

		definitionJSON, err := definition.ToJson()

		if err != nil {
			containers.Logger.Info(err.Error())
		} else {
			response, err := shared.Manager.KindsRegistry["container"].Apply(user, definitionJSON, agent)

			if err != nil {
				logger.Log.Error(err.Error())
				return
			}

			if response.Success {
				containers.Logger.Info(response.Explanation)
			} else {
				containers.Logger.Error(response.ErrorExplanation)
			}
		}
	}

	containers.Syncing = false
}
