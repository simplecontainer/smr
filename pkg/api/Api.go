package api

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/config"
	"smr/pkg/container"
	"smr/pkg/dns"
	"smr/pkg/gitops"
	"smr/pkg/manager"
	"smr/pkg/reconciler"
	"smr/pkg/registry"
	"smr/pkg/runtime"
)

func NewApi(config *config.Config, badger *badger.DB) *Api {
	api := &Api{
		Config:              config,
		Runtime:             &runtime.Runtime{},
		Registry:            &registry.Registry{},
		Reconciler:          reconciler.New(),
		Manager:             &manager.Manager{},
		Badger:              badger,
		DnsCache:            &dns.Records{},
		RepostitoryWatchers: &gitops.RepositoryWatchers{},
	}

	api.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	api.RepostitoryWatchers = &gitops.RepositoryWatchers{
		Repositories: make(map[string]*gitops.Gitops),
	}

	api.Runtime = runtime.GetRuntimeInfo()
	api.Manager.Config = api.Config
	api.Manager.Runtime = api.Runtime
	api.Manager.Registry = api.Registry
	api.Manager.Reconciler = api.Reconciler
	api.Manager.Badger = badger
	api.Manager.DnsCache = api.DnsCache
	api.Manager.RepositoryWatchers = api.RepostitoryWatchers

	return api
}
