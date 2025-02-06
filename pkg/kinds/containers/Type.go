package containers

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/static"
)

type Containers struct {
	Started bool
	Shared  *shared.Shared
}

const KIND string = static.KIND_CONTAINERS
