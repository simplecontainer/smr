package controls

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
	"time"
)

type Command interface {
	Name() string
	Time() time.Time
	Node(*manager.Manager, map[string]string) error
	Agent(*client.Client, map[string]string) error
	Data() map[string]string
	NodeID() uint64
	SetNodeID(uint64)
}

type GenericCommand struct {
	nodeID    uint64
	name      string
	data      map[string]string
	timestamp time.Time
}
