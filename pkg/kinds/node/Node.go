package node

import (
	"github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(mgr *manager.Manager) *Node {
	return &Node{
		Shared: &shared.Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
