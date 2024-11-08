package docker

import (
	DTTypes "github.com/docker/docker/api/types"
	DTEvents "github.com/docker/docker/api/types/events"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
)

func Event(event DTEvents.Message) (string, string, bool, string) {
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
		return "", "", false, ""
	}

	if err != nil {
		return "", "", false, ""
	}

	managed := false

	if c.Labels["managed"] == "smr" {
		managed = true
	}

	name := c.Labels["name"]
	group := c.Labels["group"]

	switch event.Action {
	case "connect":
		return group, name, managed, types.EVENT_NETWORK_CONNECT
	case "disconnect":
		return group, name, managed, types.EVENT_NETWORK_DISCONNECT
	case "start":
		return group, name, managed, types.EVENT_START
	case "kill":
		return group, name, managed, types.EVENT_KILL
	case "stop":
		return group, name, managed, types.EVENT_STOP
	case "die":
		return group, name, managed, types.EVENT_DIE
	default:
		return group, name, managed, ""
	}
}
