package platforms

import (
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
)

type Container struct {
	Platform IPlatform
	General  *General
	Type     string
}

type General struct {
	Labels  map[string]string
	Runtime *types.Runtime
	Status  *status.Status
}

const SPREAD_SPECIFIC string = "specific"
const SPREAD_UNIFORM string = "uniform"
