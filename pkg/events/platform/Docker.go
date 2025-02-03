package platform

import (
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
)

func DockerNew(event DTEvents.Message) Event {
	var c DTTypes.Container
	var err error

	switch event.Type {
	case DTEvents.NetworkEventType:
		c, err = docker.DockerGet(event.Actor.Attributes["container"])
		break
	case DTEvents.ContainerEventType:
		c, err = docker.DockerGet(event.Actor.ID)
		break
	default:
		return Event{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}

	if err != nil {
		return Event{
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
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_CONNECT,
		}
	case "disconnect":
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_DISCONNECT,
		}
	case "start":
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_START,
		}
	case "kill":
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_KILL,
		}
	case "stop":
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_STOP,
		}
	case "die":
		return Event{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_DIE,
		}
	default:
		return Event{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}
}
