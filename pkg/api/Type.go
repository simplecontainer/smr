package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objectdependency"
	"sync"
)

type Api struct {
	Config             *configuration.Configuration
	Keys               *keys.Keys
	DnsCache           *dns.Records
	Badger             *badger.DB
	BadgerSync         *sync.RWMutex
	DefinitionRegistry *objectdependency.DefinitionRegistry
	Manager            *manager.Manager
}

type Kv struct {
	Value string
}
