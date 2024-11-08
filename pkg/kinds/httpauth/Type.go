package httpauth

import (
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Httpauth struct {
	Started    bool
	Shared     *Shared
	Definition v1.HttpAuthDefinition
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = "httpauth"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
