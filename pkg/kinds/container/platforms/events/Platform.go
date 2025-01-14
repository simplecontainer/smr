package events

import (
	"context"
	"fmt"
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
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
	var group string
	var name string
	var event string
	var managed bool

	switch platform {
	case static.PLATFORM_DOCKER:
		group, name, managed, event = docker.Event(msg.(DTEvents.Message))
		break
	}

	containerObj := shared.Registry.FindLocal(group, name)
	if containerObj == nil {
		return
	}

	if managed {
		switch event {
		case "connect":
			HandleConnect(shared, containerObj)
		case "disconnect":
			HandleDisconnect(shared, containerObj)
			break
		case "start":
			HandleStart(shared, containerObj)
		case "kill":
			HandleKill(shared, containerObj)
		case "stop":
			HandleStop(shared, containerObj)
		case "die":
			HandleDie(shared, containerObj)
		default:

		}
	}
}

func HandleConnect(shared *shared.Shared, container platforms.IContainer) {
	logger.Log.Info(fmt.Sprintf("container is connected to the network: %s", container.GetGeneratedName()))
	container.UpdateDns(shared.DnsCache)
}

func HandleDisconnect(shared *shared.Shared, container platforms.IContainer) {
	logger.Log.Info(fmt.Sprintf("container is disconnected from the network: %s", container.GetGeneratedName()))
	container.UpdateDns(shared.DnsCache)
}

func HandleStart(shared *shared.Shared, container platforms.IContainer) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		// NO OP
	}
}

func HandleKill(shared *shared.Shared, container platforms.IContainer) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is killed - reconcile %s", container.GetGeneratedName()))
		//container.GetStatus().TransitionState(container.GetGeneratedName(), status.STATUS_KILL)
	}
}

func HandleStop(shared *shared.Shared, container platforms.IContainer) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleDie(shared *shared.Shared, container platforms.IContainer) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile %s", container.GetGeneratedName()))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}
