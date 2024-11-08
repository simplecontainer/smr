package resource

import (
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Resource struct {
	Started    bool
	Shared     *Shared
	Definition v1.ResourceDefinition
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = "resource"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
