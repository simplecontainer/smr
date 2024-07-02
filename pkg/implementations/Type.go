package implementations

import (
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Start(*manager.Manager) error
	Apply([]byte) (httpcontract.ResponseImplementation, error)
	Compare([]byte) (httpcontract.ResponseImplementation, error)
	Delete([]byte) (httpcontract.ResponseImplementation, error)
}
