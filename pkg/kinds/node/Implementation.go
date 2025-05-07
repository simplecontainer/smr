package node

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"net/http"
)

func (node *Node) Start() error {
	node.Started = true
	return nil
}
func (node *Node) GetShared() ishared.Shared {
	return node.Shared
}

func (node *Node) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	return common.Response(http.StatusOK, "object can't be applied", nil, nil), nil
}

func (node *Node) Replay(user *authentication.User) (iresponse.Response, error) {
	return iresponse.Response{}, nil
}

func (node *Node) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	return common.Response(http.StatusOK, "object can't be stated", nil, nil), nil
}

func (node *Node) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	return common.Response(http.StatusOK, "object can't be deleted", nil, nil), nil
}

func (node *Node) Event(event ievents.Event) error {
	return nil
}
