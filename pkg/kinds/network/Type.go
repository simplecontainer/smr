package network

import (
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Network struct {
	Started    bool
	Shared     *Shared
	Definition v1.NetworkDefinition
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = "network"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
