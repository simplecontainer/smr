package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/config"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objectdependency"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/runtime"
)

type Api struct {
	Config             *config.Config
	Runtime            *runtime.Runtime
	Registry           *registry.Registry
	Keys               *keys.Keys
	DnsCache           *dns.Records
	Badger             *badger.DB
	BadgerEncrypted    *badger.DB
	DefinitionRegistry *objectdependency.DefinitionRegistry
	Manager            *manager.Manager
}

type Kv struct {
	Value string
}
