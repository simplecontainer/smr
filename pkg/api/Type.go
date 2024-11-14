package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/relations"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"sync"
)

type Api struct {
	User          *authentication.User
	Config        *configuration.Configuration
	Keys          *keys.Keys
	DnsCache      *dns.Records
	Badger        *badger.DB
	confChangeC   chan raftpb.ConfChange
	Cluster       *cluster.Cluster
	BadgerSync    *sync.RWMutex
	Kinds         *relations.RelationRegistry
	KindsRegistry map[string]contracts.Kind
	Manager       *manager.Manager
	VersionServer string
}

type Kv struct {
	Value string
	Auth  string
}
