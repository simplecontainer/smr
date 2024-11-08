package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type Kind interface {
	Start() error
	Apply(*authentication.User, []byte) (ResponseImplementation, error)
	Compare(*authentication.User, []byte) (ResponseImplementation, error)
	Delete(*authentication.User, []byte) (ResponseImplementation, error)
	GetShared() interface{}
	Run(string, ...interface{}) ResponseOperator
}
