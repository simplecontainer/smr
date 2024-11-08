package containers

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
)

type Containers struct {
	Started    bool
	Shared     *shared.Shared
	Definition v1.ContainersDefinition
}

const KIND string = "containers"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
