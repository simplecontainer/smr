package shared

import (
	"github.com/simplecontainer/smr/implementations/hub/hub"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Shared struct {
	Event   chan *hub.Event
	Manager *manager.Manager
}
