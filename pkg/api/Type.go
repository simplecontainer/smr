package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/relations"
	"sync"
)

type Api struct {
	Config           *configuration.Configuration
	Keys             *keys.Keys
	DnsCache         *dns.Records
	Badger           *badger.DB
	BadgerSync       *sync.RWMutex
	RelationRegistry *relations.RelationRegistry
	Manager          *manager.Manager
	VersionServer    string
}

type Kv struct {
	Value string
}
