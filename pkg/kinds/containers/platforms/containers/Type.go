package containers

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"sync"
)

type Container struct {
	Platform platforms.IPlatform
	General  *General
	Type     string
	ghost    bool
	Lock     sync.RWMutex
}

type General struct {
	Labels  map[string]string
	Runtime *types.Runtime
	Status  *status.Status
}

const SPREAD_SPECIFIC string = "specific"
const SPREAD_UNIFORM string = "uniform"
