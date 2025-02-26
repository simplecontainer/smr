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
				go Handle(platform, shared, msg)
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

	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).Container.GetStatus().TransitionState(container.GetGroup(), container.GetGeneratedName(), status.STATUS_KILL)
		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleStart(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) {
		// NO OP YET
	}
}

func HandleKill(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) {
		logger.Log.Info(fmt.Sprintf("container is killed - event ignored till container is exited %s", container.GetGeneratedName()))
	}
}

func HandleStop(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	if !reconcileIgnore(container.GetLabels()) && container.GetStatus().GetCategory() != status.CATEGORY_END {
		logger.Log.Info(fmt.Sprintf("container is stopped - reconcile will now trigger %s", container.GetGeneratedName()))

		// wait for exited
		err := container.Wait()

		if err != nil {
			container.GetStatus().SetState(status.STATUS_DAEMON_FAILURE)
			return
		}

		if container.GetStatus().GetCategory() == status.CATEGORY_CLEAN {
			switch container.GetStatus().State.PreviousState {
			case status.STATUS_CREATED, status.STATUS_PENDING_DELETE:
				container.GetStatus().SetState(container.GetStatus().State.PreviousState)
			default:
				container.GetStatus().SetState(status.STATUS_DEAD)
			}
		} else {
			container.GetStatus().SetState(status.STATUS_DEAD)
		}

		shared.Watchers.Find(fmt.Sprintf("%s.%s", container.GetGroup(), container.GetGeneratedName())).ContainerQueue <- container
	}
}

func HandleDie(shared *shared.Shared, container platforms.IContainer, event ievents.Event) {
	// NO OP
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
