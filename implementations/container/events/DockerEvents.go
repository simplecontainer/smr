package events

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/utils"
	"go.uber.org/zap"
)

func ListenDockerEvents(shared *shared.Shared) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cEvents, cErr := cli.Events(ctx, types.EventsOptions{})

	for {
		select {
		case err := <-cErr:
			fmt.Println(err)
		case msg := <-cEvents:
			// TODO: Do I want to do blocking here? Or go with gouroutine?
			HandleDockerEvent(shared, msg)
		}
	}
}

func HandleDockerEvent(shared *shared.Shared, event events.Message) {
	var container *container.Container

	// fetch container info from container events
	if utils.Contains([]string{"start", "kill", "stop", "die"}, event.Action) {
		container = shared.Registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
	}

	// fetch container info from network events
	if utils.Contains([]string{"connect", "disconnect"}, event.Action) {
		c := container.GetFromId(event.Actor.Attributes["container"])
		fmt.Println(c)

		container = shared.Registry.Find(c.Labels["group"], c.Labels["name"])
	}

	if container == nil {
		fmt.Println("container not found in registry")
		return
	}

	c := container.Get()

	if c == nil {
		return
	}

	managed := false

	// only manage smr created containers, others are left alone to live and die in peace
	if c.Labels["managed"] == "smr" {
		managed = true
	}

	switch event.Action {
	case "connect":
		if managed {
			HandleConnect(shared, container, event)
		}
		break
	case "disconnect":
		if managed {
			HandleDisconnect(shared, container, event)
		}
		break
	case "start":
		if managed {
			HandleStart(shared, container, event)
		}
		break
	case "kill":
		if managed {
			HandleKill(shared, container, event)
		}
		break
	case "stop":
		if managed {
			HandleStop(shared, container, event)
		}
		break
	case "die":
		if managed {
			HandleDie(shared, container, event)
		}
		break
	default:
	}
}

func HandleConnect(shared *shared.Shared, container *container.Container, event events.Message) {
	// Handle network connect here
}

func HandleDisconnect(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	for _, ip := range shared.DnsCache.FindDeleteQueue(containerObj.GetDomain()) {
		shared.DnsCache.RemoveARecord(containerObj.GetDomain(), ip)
		shared.DnsCache.RemoveARecord(containerObj.GetHeadlessDomain(), ip)
	}

	shared.DnsCache.ResetDeleteQueue(containerObj.GetDomain())
}

func HandleStart(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	// Container started it is running so update status accordingly
	// containerObj.Status.TransitionState(status.STATUS_RUNNING)
}

func HandleKill(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	// It can happen that kill signal occurs in the container even if it is not dying; eg killing thread, goroutine etc.
	containerObj.Status.TransitionState(status.STATUS_KILLED)

	for _, n := range containerObj.Runtime.Networks {
		shared.DnsCache.RemoveARecordQueue(containerObj.GetDomain(), n.IP)
	}
}

func HandleStop(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	// Stop will stop the container so update the status accordingly
	// containerObj.Status.TransitionState(status.STATUS_DEAD)
}

func HandleDie(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	fmt.Println("Container died")
	fmt.Println(containerObj.Status.GetState())
	containerObj.Status.TransitionState(status.STATUS_DEAD)

	reconcile := true

	// labels for ignoring events for specific container
	val, exists := containerObj.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", containerObj.Static.GeneratedName))
			reconcile = false
		}
	}

	if reconcile {
		logger.Log.Info(fmt.Sprintf("container is dead - reconcile %s", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
	}
}
