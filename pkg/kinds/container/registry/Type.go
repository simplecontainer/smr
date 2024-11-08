package registry

import (
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/objects"
	"sync"
)

type Registry struct {
	Containers     map[string]map[string]platforms.IContainer
	ContainersLock sync.RWMutex
	Indexes        map[string][]int
	BackOffTracker map[string]map[string]int
	Object         map[string]objects.Object
}
