package gitops

import (
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(mgr *manager.Manager) *Gitops {
	return &Gitops{
		Shared: &shared.Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
