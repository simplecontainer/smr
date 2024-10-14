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
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
)

func ListenDockerEvents(shared *shared.Shared) {
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
	var c *types.Container

	if event.Actor.Attributes["container"] != "" {
		c = container.GetFromId(event.Actor.Attributes["container"])
	} else {
		c = container.GetFromId(event.ID)
	}

	if c == nil {
		logger.Log.Debug("container is not found in the docker daemon",
			zap.String("container", event.Actor.Attributes["container"]),
		)

		return
	}

	container = shared.Registry.Find(c.Labels["group"], c.Labels["name"])

	if container == nil {
		logger.Log.Debug("container is not found in the registry, ignore it",
			zap.String("container", c.Names[0]),
		)

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
	for _, n := range containerObj.Runtime.Networks.Networks {
		for _, ip := range shared.DnsCache.FindDeleteQueue(containerObj.GetDomain(n.Reference.Name)) {
			shared.DnsCache.RemoveARecord(containerObj.GetDomain(n.Reference.Name), ip)
			shared.DnsCache.RemoveARecord(containerObj.GetHeadlessDomain(n.Reference.Name), ip)

			obj := objects.New(shared.Client.Get("root"), &authentication.User{
				Username: "root",
				Domain:   "localhost",
			})

			obj.Remove(f.NewFromString(fmt.Sprintf("network.%s.%s.dns", containerObj.Static.Group, containerObj.Static.GeneratedName)))
		}

		shared.DnsCache.ResetDeleteQueue(containerObj.GetDomain(n.Reference.Name))
	}
}

func HandleStart(shared *shared.Shared, containerObj *container.Container, event events.Message) {
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
		logger.Log.Info(fmt.Sprintf("container is stopped- reconcile %s", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
	}
}

func HandleKill(shared *shared.Shared, containerObj *container.Container, event events.Message) {
	containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_KILL)
}

func HandleStop(shared *shared.Shared, containerObj *container.Container, event events.Message) {
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
		logger.Log.Info(fmt.Sprintf("container is stopped- reconcile %s", containerObj.Static.GeneratedName))
		shared.Watcher.Find(fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)).ContainerQueue <- containerObj
	}
}

func HandleDie(shared *shared.Shared, containerObj *container.Container, event events.Message) {
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
