package ikinds

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
)

type Kind interface {
	Start() error
	Apply(*authentication.User, []byte, string) (iresponse.Response, error)
	State(*authentication.User, []byte, string) (iresponse.Response, error)
	Delete(*authentication.User, []byte, string) (iresponse.Response, error)
	GetShared() interface{}
	Event(events ievents.Event) error
}
