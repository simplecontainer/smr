package network

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Network {
	return &Network{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
