package listener

import (
	"context"
	"fmt"
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
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

		cEvents, cErr := cli.Events(ctx, DTTypes.EventsOptions{})

		for {
			select {
			case err = <-cErr:
				logger.Log.Error(err.Error())
			case msg := <-cEvents:
				// TODO: Do I want to do blocking here? Or go with gouroutine?
				Handle(platform, shared, msg)
			}
		}
	}
}

func Handle(platform string, shared *shared.Shared, msg interface{}) {
	var event ievents.Event
	switch platform {
	case static.PLATFORM_DOCKER:
		event = docker.NewEvent(msg.(DTEvents.Message))
		break
	}

	var containerObj platforms.IContainer

	if event.GetGroup() != "" && event.GetName() != "" {
		containerObj = shared.Registry.FindLocal(event.GetGroup(), event.GetName())
	}

	if containerObj == nil {
		return
	}

	if event.IsManaged() {
		switch event.GetType() {
		case types.EVENT_NETWORK_CONNECT:
			HandleConnect(shared, containerObj, event)
			return
		case types.EVENT_NETWORK_DISCONNECT:
			HandleDisconnect(shared, containerObj, event)
			return
		case types.EVENT_START:
			HandleStart(shared, containerObj, event)
			return
		case types.EVENT_KILL:
			HandleKill(shared, containerObj, event)
			return
		case types.EVENT_STOP:
			HandleStop(shared, containerObj, event)
			return
		case types.EVENT_DIE:
			HandleDie(shared, containerObj, event)
			return
		default:
			return
		}
	}
}

func HandleConnect(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	logger.Log.Info(fmt.Sprintf("container %s is connected to the network: %s", container.GetGeneratedName(), event.GetNetworkId()))
	err := container.SyncNetwork()

	if err != nil {
		logger.Log.Error(err.Error())
	}

	err = container.UpdateDns(shared.DnsCache)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func HandleDisconnect(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	logger.Log.Info(fmt.Sprintf("container %s is disconnected from the network: %s", container.GetGeneratedName(), event.GetNetworkId()))
	err := container.RemoveDns(shared.DnsCache, event.GetNetworkId())

	if err != nil {
		logger.Log.Error(err.Error())
	}

	err = container.SyncNetwork()

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func HandleStart(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) {
		// NO OP
	}
}

func HandleKill(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) {
		logger.Log.Info(fmt.Sprintf("container is killed - event ignored till container is exited %s", container.GetGeneratedName()))
	}
}

func HandleStop(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	// NO OP
}

func HandleDie(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) {
		containerW := shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName()))

		if containerW.AllowPlatformEvents {
			logger.Log.Info(fmt.Sprintf("container is stopped - reconcile to dead %s", container.GetGeneratedName()))

			container.GetStatus().GetPending().Clear()
			container.GetStatus().SetState(status.DEAD)

			shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
		} else {
			logger.Log.Info(fmt.Sprintf("container is stopped - reconcile will be ignored since it is not allowed %s", container.GetGeneratedName()))
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
