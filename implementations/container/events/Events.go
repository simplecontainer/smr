package events

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/qdnqn/smr/implementations/container/container"
	"github.com/qdnqn/smr/implementations/container/status"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/utils"
	"go.uber.org/zap"
)

func ListenDockerEvents(mgr *manager.Manager) {
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
			DockerEvent(mgr, msg)
		}
	}
}

func ListenEvents(mgr *manager.Manager, reconcilerObj reconciler.Reconciler) {
	for {
		select {
		case event := <-reconcilerObj.QueueEvents:
			Event(mgr, event)
		}
	}
}

func Event(mgr *manager.Manager, event reconciler.Events) {
	var container *container.Container

	// handle container events
	if utils.Contains([]string{"change"}, event.Kind) {
		container = mgr.Registry.Find(event.Container.Static.Group, event.Container.Static.GeneratedName)
	}

	if container == nil {
		return
	}

	c := container.Get()
	managed := false

	// only manage smr created containers, others are left alone to live and die in peace
	if c.Labels["managed"] == "smr" {
		managed = true
	}

	switch event.Kind {
	case "change":
		if managed {
			HandleChange(mgr, container)
		}
		break
	default:
	}
}

func HandleChange(mgr *manager.Manager, containerObj *container.Container) {
	reconcile := true

	// labels for ignoring events for specific container
	val, exists := containerObj.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", containerObj.Static.GeneratedName))
			reconcile = false
		}
	}

	if !containerObj.Status.TransitionState(status.STATUS_RECONCILING) && reconcile {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", containerObj.Static.GeneratedName))
		containerObj.Status.TransitionState("created")
		mgr.ContainerWatchers.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj
	}
}

func DockerEvent(mgr *manager.Manager, event events.Message) {
	var container *container.Container

	// handle container events
	if utils.Contains([]string{"start", "kill", "stop", "die"}, event.Action) {
		container = mgr.Registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
	}

	// handle network events
	if utils.Contains([]string{"connect", "disconnect"}, event.Action) {
		c := container.GetFromId(event.Actor.Attributes["container"])
		container = mgr.Registry.Find(c.Labels["group"], c.Labels["name"])
	}

	if container == nil {
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
			HandleConnect(mgr, container, event)
		}
		break
	case "disconnect":
		if managed {
			HandleDisconnect(mgr, container, event)
		}
		break
	case "start":
		if managed {
			HandleStart(mgr, container, event)
		}
		break
	case "kill":
		if managed {
			HandleKill(mgr, container, event)
		}
		break
	case "stop":
		if managed {
			HandleStop(mgr, container, event)
		}
		break
	case "die":
		if managed {
			HandleDie(mgr, container, event)
		}
		break
	default:
	}
}

func HandleConnect(mgr *manager.Manager, container *container.Container, event events.Message) {
	// Handle network connect here
}

func HandleDisconnect(mgr *manager.Manager, containerObj *container.Container, event events.Message) {
	for _, ip := range mgr.DnsCache.FindDeleteQueue(containerObj.GetDomain()) {
		mgr.DnsCache.RemoveARecord(containerObj.GetDomain(), ip)
		mgr.DnsCache.RemoveARecord(containerObj.GetHeadlessDomain(), ip)
	}

	mgr.DnsCache.ResetDeleteQueue(containerObj.GetDomain())
}

func HandleStart(mgr *manager.Manager, containerObj *container.Container, event events.Message) {
	// Container started it is running so update status accordingly
	// containerObj.Status.TransitionState(status.STATUS_RUNNING)
}

func HandleKill(mgr *manager.Manager, containerObj *container.Container, event events.Message) {
	// It can happen that kill signal occurs in the container even if it is not dying; eg killing thread, goroutine etc.
	containerObj.Status.TransitionState(status.STATUS_KILLED)

	for _, n := range containerObj.Runtime.Networks {
		mgr.DnsCache.RemoveARecordQueue(containerObj.GetDomain(), n.IP)
	}
}

func HandleStop(mgr *manager.Manager, containerObj *container.Container, event events.Message) {
	// Stop will stop the container so update the status accordingly
	// containerObj.Status.TransitionState(status.STATUS_DEAD)
}

func HandleDie(mgr *manager.Manager, containerObj *container.Container, event events.Message) {
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
		mgr.ContainerWatchers.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
	}
}
