package api

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/config"
	"smr/pkg/dns"
	"smr/pkg/gitops"
	"smr/pkg/manager"
	"smr/pkg/reconciler"
	"smr/pkg/registry"
	"smr/pkg/runtime"
)

type Api struct {
	Config              *config.Config
	Runtime             *runtime.Runtime
	Registry            *registry.Registry
	Reconciler          *reconciler.Reconciler
	Manager             *manager.Manager
	Badger              *badger.DB
	DnsCache            *dns.Records
	RepostitoryWatchers *gitops.RepositoryWatcher
}

type Kv struct {
	Value string
}
