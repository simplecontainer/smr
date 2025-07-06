package volume

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Volume {
	return &Volume{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
