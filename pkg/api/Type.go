package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/config"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objectdependency"
	"github.com/simplecontainer/smr/pkg/registry"
	"github.com/simplecontainer/smr/pkg/runtime"
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
