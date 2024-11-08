package container

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
)

type Container struct {
	Started bool
	Shared  *shared.Shared
}

type Request struct {
	Definition *v1.ContainerDefinition
}

const KIND string = "container"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
