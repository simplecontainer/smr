package docker

import (
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
)

func Event(event DTEvents.Message) contracts.PlatformEvent {
	var c DTTypes.Container
	var err error

	switch event.Type {
	case DTEvents.NetworkEventType:
		c, err = DockerGet(event.Actor.Attributes["container"])
		break
	case DTEvents.ContainerEventType:
		c, err = DockerGet(event.Actor.ID)
		break
	default:
		return contracts.PlatformEvent{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}

	if err != nil {
		return contracts.PlatformEvent{
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
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_CONNECT,
		}
	case "disconnect":
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_NETWORK_DISCONNECT,
		}
	case "start":
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_START,
		}
	case "kill":
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_KILL,
		}
	case "stop":
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_STOP,
		}
	case "die":
		return contracts.PlatformEvent{
			NetworkID:   event.Actor.ID,
			ContainerID: c.ID,
			Group:       group,
			Name:        name,
			Managed:     managed,
			Type:        types.EVENT_DIE,
		}
	default:
		return contracts.PlatformEvent{
			NetworkID:   "",
			ContainerID: "",
			Group:       "",
			Name:        "",
			Managed:     false,
			Type:        "unknown",
		}
	}
}
