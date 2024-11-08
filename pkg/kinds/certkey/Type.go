package certkey

import (
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/manager"
)

type Certkey struct {
	Started    bool
	Shared     *Shared
	Definition v1.CertKeyDefinition
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
