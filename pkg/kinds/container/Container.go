package container

import (
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(mgr *manager.Manager) *Container {
	return &Container{
		Started: false,
		Shared: &shared.Shared{
			Manager:  mgr,
			Client:   mgr.Http,
			DnsCache: mgr.DnsCache,
		},
	}
}
