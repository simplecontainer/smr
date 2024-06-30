package reconciler

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/status"
	"github.com/qdnqn/smr/pkg/utils"
	"go.uber.org/zap"
)

func (reconciler *Reconciler) ListenDockerEvents(registry *registry.Registry, dnsCache *dns.Records) {
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
			reconciler.DockerEvent(registry, dnsCache, msg)
		}
	}
}

func (reconciler *Reconciler) ListenEvents(registry *registry.Registry, dnsCache *dns.Records) {
	for {
		select {
		case event := <-reconciler.QueueEvents:
			reconciler.Event(registry, dnsCache, event)
		}
	}
}

func (reconciler *Reconciler) DockerEvent(registry *registry.Registry, dnsCache *dns.Records, event events.Message) {
	var container *container.Container

	// handle container events
	if utils.Contains([]string{"start", "kill", "stop", "die"}, event.Action) {
		container = registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
	}

	// handle network events
	if utils.Contains([]string{"connect", "disconnect"}, event.Action) {
		c := container.GetFromId(event.Actor.Attributes["container"])
		container = registry.Find(c.Labels["group"], c.Labels["name"])
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

	switch event.Action {
	case "connect":
		if managed {
			reconciler.HandleConnect(registry, dnsCache, container, event)
		}
		break
	case "disconnect":
		if managed {
			reconciler.HandleDisconnect(registry, dnsCache, container, event)
		}
		break
	case "start":
		if managed {
			reconciler.HandleStart(registry, container)
		}
		break
	case "kill":
		if managed {
			reconciler.HandleKill(registry, dnsCache, container)
		}
		break
	case "stop":
		if managed {
			reconciler.HandleStop(registry, container)
		}
		break
	case "die":
		if managed {
			reconciler.HandleDie(registry, container)
		}
		break
	default:
	}
}

func (reconciler *Reconciler) HandleConnect(registry *registry.Registry, dnsCache *dns.Records, container *container.Container, event events.Message) {
	// Handle network connect here
}

func (reconciler *Reconciler) HandleDisconnect(registry *registry.Registry, dnsCache *dns.Records, container *container.Container, event events.Message) {
	for _, ip := range dnsCache.FindDeleteQueue(container.GetDomain()) {
		dnsCache.RemoveARecord(container.GetDomain(), ip)
		dnsCache.RemoveARecord(container.GetHeadlessDomain(), ip)
	}

	dnsCache.ResetDeleteQueue(container.GetDomain())
}

func (reconciler *Reconciler) HandleStart(registry *registry.Registry, container *container.Container) {
	// Container started it is running so update status accordingly
	container.Status.TransitionState(status.STATUS_RUNNING)
}

func (reconciler *Reconciler) HandleKill(registry *registry.Registry, dnsCache *dns.Records, container *container.Container) {
	// It can happen that kill signal occurs in the container even if it is not dying; eg killing thread, goroutine etc.
	container.Status.TransitionState(status.STATUS_KILLED)

	for _, n := range container.Runtime.Networks {
		dnsCache.RemoveARecordQueue(container.GetDomain(), n.IP)
	}
}

func (reconciler *Reconciler) HandleStop(registry *registry.Registry, container *container.Container) {
	// Stop will stop the container so update the status accordingly
	container.Status.TransitionState(status.STATUS_DEAD)
}

func (reconciler *Reconciler) HandleDie(registry *registry.Registry, container *container.Container) {
	container.Status.TransitionState(status.STATUS_DEAD)

	reconcile := true

	// labels for ignoring events for specific container
	val, exists := container.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", container.Static.GeneratedName))
			reconcile = false
		}
	}

	if container.Status.IfStateIs(status.STATUS_DEAD) && reconcile {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", container.Static.GeneratedName))
		reconciler.QueueChan <- Reconcile{
			Container: container,
		}
	}
}
