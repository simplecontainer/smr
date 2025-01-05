package resource

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Resource struct {
	Started bool
	Shared  *Shared
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
