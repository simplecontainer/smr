package listener

import (
	"context"
	"fmt"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/metrics"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"time"
)

func Listen(shared *shared.Shared, platform string) {
	switch platform {
	case static.PLATFORM_DOCKER:
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

		if err != nil {
			panic(err)
		}

		defer func(cli *client.Client) {
			err = cli.Close()
			if err != nil {
				return
			}
		}(cli)

		cEvents, cErr := cli.Events(ctx, DTEvents.ListOptions{})

		for {
			select {
			case err = <-cErr:
				logger.Log.Error(err.Error())
			case msg := <-cEvents:
				go Handle(platform, shared, msg)
			}
		}
	}
}

func Handle(platform string, shared *shared.Shared, msg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("panic in platform event handler", zap.Any("panic", r))
		}
	}()

	var event ievents.Event
	switch platform {
	case static.PLATFORM_DOCKER:
		event = docker.NewEvent(msg.(DTEvents.Message))
		break
	}

	var cw *watcher.Container

	if event.GetGroup() != "" && event.GetName() != "" {
		cw = shared.Watchers.Find(fmt.Sprintf("%s/%s", event.GetGroup(), event.GetName()))
	}

	if cw == nil || cw.IsDone() {
		return
	}

	if event.IsManaged() {
		switch event.GetType() {
		case types.EVENT_NETWORK_CONNECT:
			HandleConnect(shared, cw, event)
			return
		case types.EVENT_NETWORK_DISCONNECT:
			HandleDisconnect(shared, cw, event)
			return
		case types.EVENT_START:
			HandleStart(shared, cw, event)
			return
		case types.EVENT_KILL:
			HandleKill(shared, cw, event)
			return
		case types.EVENT_STOP:
			HandleStop(shared, cw, event)
			return
		case types.EVENT_DIE:
			HandleDie(shared, cw, event)
			return
		default:
			return
		}
	}
}
func HandleConnect(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	cw.Logger.Info(fmt.Sprintf("container %s is connected to the network: %s", cw.Container.GetGeneratedName(), event.GetNetworkId()))
	err := cw.Container.SyncNetwork()

	if err != nil {
		cw.Logger.Error(err.Error())
	}

	err = cw.Container.UpdateDns(shared.DnsCache)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func HandleDisconnect(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	cw.Logger.Info(fmt.Sprintf("container %s is disconnected from the network: %s", cw.Container.GetGeneratedName(), event.GetNetworkId()))
	err := cw.Container.RemoveDns(shared.DnsCache, event.GetNetworkId())

	if err != nil {
		cw.Logger.Error(err.Error())
	}

	err = cw.Container.SyncNetwork()

	if err != nil {
		cw.Logger.Error(err.Error())
	}
}

func HandleStart(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	if !reconcileIgnore(cw.Container.GetLabels()) {
		// NO OP
	}
}

func HandleKill(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	if !reconcileIgnore(cw.Container.GetLabels()) {
		cw.Logger.Info(fmt.Sprintf("container is killed - event ignored till container is exited %s", cw.Container.GetGeneratedName()))
	}
}

func HandleStop(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	// NO OP
}

func HandleDie(shared *shared.Shared, cw *watcher.Container, event ievents.Event) {
	if !reconcileIgnore(cw.Container.GetLabels()) {
		containerW := shared.Watchers.Find(cw.Container.GetGroupIdentifier())

		if containerW != nil && !containerW.IsDone() && containerW.GetAllowPlatformEvents() {
			cw.Logger.Info(fmt.Sprintf("container is stopped - reconcile to dead %s", cw.Container.GetGeneratedName()))

			cw.Container.GetStatus().RejectQueueAttempts(time.Now())
			cw.Container.GetStatus().GetPending().Clear()
			cw.Container.GetStatus().QueueState(status.DEAD, time.Now())

			metrics.Containers.Get().DeletePartialMatch(prometheus.Labels{"container": cw.Container.GetGeneratedName()})
			metrics.Containers.Set(1, cw.Container.GetGeneratedName(), status.DEAD)
			metrics.ContainersHistory.Set(1, cw.Container.GetGeneratedName(), status.DEAD)

			containerW.SendToQueue(cw.Container, 5*time.Second)
		} else {
			cw.Logger.Info(fmt.Sprintf("container is stopped - reconcile will be ignored since it is not allowed %s", cw.Container.GetGeneratedName()))
		}
	}
}

func reconcileIgnore(labels map[string]string) bool {
	val, exists := labels["reconcile"]

	if exists {
		if val == "false" {
			return true
		}
	}

	return false
}
