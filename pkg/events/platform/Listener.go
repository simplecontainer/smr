package platform

import (
	"context"
	"fmt"
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/contracts"
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
	var event contracts.PlatformEvent
	switch platform {
	case static.PLATFORM_DOCKER:
		event = docker.NewEvent(msg.(DTEvents.Message))
		break
	}

	containerObj := shared.Registry.FindLocal(event.Group, event.Name)
	if containerObj == nil {
		return
	}

	if event.Managed {
		switch event.Type {
		case types.EVENT_NETWORK_CONNECT:
			HandleConnect(shared, containerObj, event)
		case types.EVENT_NETWORK_DISCONNECT:
			HandleDisconnect(shared, containerObj, event)
			break
		case types.EVENT_START:
			HandleStart(shared, containerObj, event)
		case types.EVENT_KILL:
			HandleKill(shared, containerObj, event)
		case types.EVENT_STOP:
			HandleStop(shared, containerObj, event)
		case types.EVENT_DIE:
			HandleDie(shared, containerObj, event)
		default:

		}
	}
}

func HandleConnect(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	logger.Log.Info(fmt.Sprintf("container %s is connected to the network: %s", container.GetGeneratedName(), event.NetworkID))
	err := container.SyncNetwork()

	if err != nil {
		logger.Log.Error(err.Error())
	}

	err = container.UpdateDns(shared.DnsCache)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func HandleDisconnect(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	logger.Log.Info(fmt.Sprintf("container %s is disconnected from the network: %s", container.GetGeneratedName(), event.NetworkID))
	err := container.SyncNetwork()

	if err != nil {
		logger.Log.Error(err.Error())
	}

	err = container.RemoveDns(shared.DnsCache, event.NetworkID)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).Container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), status.STATUS_KILL)
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleStart(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END &&
		!container.GetStatus().Reconciling {
		container.GetStatus().Recreated = false
	}
}

func HandleKill(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END &&
		!container.GetStatus().Reconciling {
		logger.Log.Info(fmt.Sprintf("container is killed - reconcile %s", container.GetGeneratedName()))
		container.GetStatus().Recreated = false
		//container.GetStatus().TransitionState(container.GetGeneratedName(), status.STATUS_KILL)
	}
}

func HandleStop(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END &&
		!container.GetStatus().Reconciling {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		container.GetStatus().Recreated = false
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleDie(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END &&
		!container.GetStatus().Reconciling {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		container.GetStatus().Recreated = false
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
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
