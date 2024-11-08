package resource

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Resource {
	return &Resource{Shared: &Shared{
		Manager: mgr,
		Client:  mgr.Http,
	}}
}
