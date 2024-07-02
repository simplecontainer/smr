package implementations

import (
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Start(*manager.Manager) error
	Apply([]byte) (httpcontract.ResponseImplementation, error)
	Compare([]byte) (httpcontract.ResponseImplementation, error)
	Delete([]byte) (httpcontract.ResponseImplementation, error)
}
