package containers

import (
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(mgr *manager.Manager) *Containers {
	return &Containers{
		Started: false,
		Shared: &shared.Shared{
			Manager:  mgr,
			Client:   mgr.Http,
			DnsCache: mgr.DnsCache,
			User:     mgr.User,
		},
	}
}
