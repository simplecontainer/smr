package operators

import (
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) httpcontract.ResponseOperator
}

type Request struct {
	Manager *manager.Manager
	Client  *http.Client
	Data    map[string]any
}
