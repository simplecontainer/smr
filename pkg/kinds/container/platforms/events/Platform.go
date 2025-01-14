package events

import (
	"context"
	"fmt"
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
)

var listeners map[string]string = make(map[string]string)

func NewPlatformEventsListener(shared *shared.Shared, platform string) {
	_, ok := listeners[platform]

	if !ok {
		listeners[platform] = platform

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
}

func Handle(platform string, shared *shared.Shared, msg interface{}) {
	var event contracts.PlatformEvent
	switch platform {
	case static.PLATFORM_DOCKER:
		event = docker.Event(msg.(DTEvents.Message))
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
	container.UpdateDns(shared.DnsCache, event.NetworkID)
}

func HandleDisconnect(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	logger.Log.Info(fmt.Sprintf("container %s is disconnected from the network: %s", container.GetGeneratedName(), event.NetworkID))
	container.RemoveDns(shared.DnsCache, event.NetworkID)
}

func HandleStart(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		// NO OP
	}
}

func HandleKill(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is killed - reconcile %s", container.GetGeneratedName()))
		//container.GetStatus().TransitionState(container.GetGeneratedName(), status.STATUS_KILL)
	}
}

func HandleStop(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleDie(shared *shared.Shared, container platforms.IContainer, event contracts.PlatformEvent) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}
