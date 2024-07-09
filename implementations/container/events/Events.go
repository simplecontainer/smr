package events

import (
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/implementations/hub/hub"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
)

func ListenEvents(shared *shared.Shared, e chan *hub.Event) {
	for {
		select {
		case event := <-e:
			Event(shared, event)
		}
	}
}

func Event(shared *shared.Shared, event *hub.Event) {
	var container *container.Container

	// handle container events
	if helpers.Contains([]string{"Container"}, event.Kind) {
		container = shared.Registry.Find(event.Group, event.Identifier)
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
			HandleChange(shared, container)
		}
		break
	default:
	}
}

func HandleChange(shared *shared.Shared, containerObj *container.Container) {
	reconcile := true

	// labels for ignoring events for specific container
	val, exists := containerObj.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", containerObj.Static.GeneratedName))
			reconcile = false
		}
	}

	if !containerObj.Status.IfStateIs(status.STATUS_RECONCILING) && reconcile {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", containerObj.Static.GeneratedName))
		containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_CREATED)
		shared.Watcher.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj
	}
}
