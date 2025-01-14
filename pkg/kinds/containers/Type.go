package containers

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/static"
)

type Containers struct {
	Started    bool
	Shared     *shared.Shared
	Definition v1.ContainersDefinition
}

const KIND string = static.KIND_CONTAINERS

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
