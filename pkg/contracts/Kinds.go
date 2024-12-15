package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type Kind interface {
	Start() error
	Apply(*authentication.User, []byte) (Response, error)
	Compare(*authentication.User, []byte) (Response, error)
	Delete(*authentication.User, []byte) (Response, error)
	GetShared() interface{}
	Run(string, Control) Response
}
