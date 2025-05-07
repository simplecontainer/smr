package ikinds

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	_ "github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
)

type Kind interface {
	Start() error
	Apply(*authentication.User, []byte, string) (iresponse.Response, error)
	Replay(*authentication.User) (iresponse.Response, error)
	State(*authentication.User, []byte, string) (iresponse.Response, error)
	Delete(*authentication.User, []byte, string) (iresponse.Response, error)
	GetShared() ishared.Shared
	Event(events ievents.Event) error
}
