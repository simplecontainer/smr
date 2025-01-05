package certkey

import (
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Certkey struct {
	Started bool
	Shared  *Shared
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = "certkey"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}
