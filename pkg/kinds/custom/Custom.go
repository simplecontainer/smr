package custom

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Custom {
	return &Custom{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
