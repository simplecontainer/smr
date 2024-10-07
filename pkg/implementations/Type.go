package implementations

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Start(*manager.Manager) error
	Apply(*authentication.User, []byte) (httpcontract.ResponseImplementation, error)
	Compare(*authentication.User, []byte) (httpcontract.ResponseImplementation, error)
	Delete(*authentication.User, []byte) (httpcontract.ResponseImplementation, error)
	GetShared() interface{}
}
