package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type Kind interface {
	Start() error
	Propose(*authentication.User, []byte, string) (Response, error)
	Apply(*authentication.User, []byte, string) (Response, error)
	Compare(*authentication.User, []byte) (Response, error)
	Delete(*authentication.User, []byte, string) (Response, error)
	GetShared() interface{}
	Run(string, Control) Response
}
