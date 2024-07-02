package operators

import (
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
)

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) httpcontract.ResponseOperator
}

type Request struct {
	Manager *manager.Manager
	Data    map[string]any
}
