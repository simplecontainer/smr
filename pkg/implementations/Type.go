package implementations

import (
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Implementation(*manager.Manager, []byte) (httpcontract.ResponseImplementation, error)
}
