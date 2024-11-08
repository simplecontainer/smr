package config

import "github.com/simplecontainer/smr/pkg/manager"

func New(mgr *manager.Manager) *Config {
	return &Config{
		Shared: &Shared{
			Manager: mgr,
			Client:  mgr.Http,
		},
	}
}
