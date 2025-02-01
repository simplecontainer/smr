package container

import (
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/static"
)

type Container struct {
	Started bool
	Shared  *shared.Shared
}

const KIND string = static.KIND_CONTAINER
