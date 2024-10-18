package httpcontract

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
)

type ResponseOperator struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             map[string]any
}

type RequestOperator struct {
	Manager *manager.Manager
	Client  *client.Http
	User    *authentication.User
	Data    map[string]any
}
