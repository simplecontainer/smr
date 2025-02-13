package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type Kind interface {
	Start() error
	Apply(*authentication.User, []byte, string) (Response, error)
	Delete(*authentication.User, []byte, string) (Response, error)
	GetShared() interface{}
	Event(events Event) error
}
