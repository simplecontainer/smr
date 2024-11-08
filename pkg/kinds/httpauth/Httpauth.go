package httpauth

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Httpauth {
	return &Httpauth{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
