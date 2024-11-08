package certkey

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Certkey {
	return &Certkey{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
