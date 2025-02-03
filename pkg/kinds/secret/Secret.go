package secret

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Secret {
	return &Secret{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
