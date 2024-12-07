package registry

import (
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/objects"
	"sync"
)

type Registry struct {
	Containers     map[string]map[string]platforms.IContainer
	ContainersLock sync.RWMutex
	Indexes        map[string][]uint64
	BackOffTracker map[string]map[string]uint64
	Object         map[string]objects.Object
}
