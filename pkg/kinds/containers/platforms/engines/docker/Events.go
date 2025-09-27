package docker

import (
	TDContainer "github.com/docker/docker/api/types/container"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/simplecontainer/smr/pkg/events/platform"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker/internal"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
)

func NewEvent(event DTEvents.Message) platform.Event {
	var c TDContainer.Summary
	var err error

	switch event.Type {
	case DTEvents.NetworkEventType:
		c, err = internal.Get(event.Actor.Attributes["container"])
		break
	case DTEvents.ContainerEventType:
		c, err = internal.Get(event.Actor.ID)
		break
	default:
		return platform.Event{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}

	if err != nil {
		return platform.Event{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}

	managed := false

	if c.Labels["managed"] == "smr" {
		managed = true
	}

	name := c.Labels["name"]
	group := c.Labels["group"]

	switch event.Action {
	case "connect":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_CONNECT,
		}
	case "disconnect":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_DISCONNECT,
		}
	case "start":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_START,
		}
	case "kill":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_KILL,
		}
	case "stop":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_STOP,
		}
	case "die":
		return platform.Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_DIE,
		}
	default:
		return platform.Event{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}
}
