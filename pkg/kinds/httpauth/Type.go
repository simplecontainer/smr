package httpauth

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
)

type Httpauth struct {
	Started bool
	Shared  *Shared
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = static.KIND_HTTPAUTH

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
