package implementations

import (
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

// Plugin contracts
type Implementation interface {
	Start(*manager.Manager) error
	Apply([]byte) (httpcontract.ResponseImplementation, error)
	Compare([]byte) (httpcontract.ResponseImplementation, error)
	Delete([]byte) (httpcontract.ResponseImplementation, error)
	GetShared() interface{}
}

type ImplementationLight interface {
	Start() error
	Apply([]byte) (httpcontract.ResponseImplementation, error)
	Compare([]byte) (httpcontract.ResponseImplementation, error)
	Delete([]byte) (httpcontract.ResponseImplementation, error)
	GetShared() interface{}
}

type ImplementationShared interface {
	GetClient(keys *keys.Keys) http.Client
}
