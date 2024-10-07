package operators

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
)

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) httpcontract.ResponseOperator
}

type Request struct {
	Manager *manager.Manager
	Client  *client.Http
	User    *authentication.User
	Data    map[string]any
}
