package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/config"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/gitops"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/runtime"
)

func NewApi(config *config.Config, badger *badger.DB) *Api {
	api := &Api{
		Config:              config,
		Runtime:             &runtime.Runtime{},
		Registry:            &registry.Registry{},
		Reconciler:          reconciler.New(),
		Keys:                &keys.Keys{},
		RepostitoryWatchers: &gitops.RepositoryWatcher{},
		DnsCache:            &dns.Records{},
		Badger:              badger,
		Manager:             &manager.Manager{},
	}

	api.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	api.RepostitoryWatchers.Repositories = make(map[string]*gitops.Gitops)

	api.Runtime = runtime.GetRuntimeInfo()
	api.Manager.Config = api.Config
	api.Manager.Runtime = api.Runtime
	api.Manager.Registry = api.Registry
	api.Manager.Reconciler = api.Reconciler
	api.Manager.Keys = api.Keys
	api.Manager.Badger = badger
	api.Manager.DnsCache = api.DnsCache
	api.Manager.RepositoryWatchers = api.RepostitoryWatchers

	return api
}
