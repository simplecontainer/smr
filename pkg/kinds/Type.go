package kinds

import (
	"github.com/simplecontainer/smr/pkg/httpcontract"
)

// Operator contracts
type Operator interface {
	Run(string, ...interface{}) httpcontract.ResponseOperator
}
