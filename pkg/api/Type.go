package api

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/wss"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/embed"
)

type Api struct {
	Server          *embed.Etcd
	Etcd            *clientv3.Client
	LeaseIdentifier *clientv3.LeaseGrantResponse
	User            *authentication.User
	Config          *configuration.Configuration
	Keys            *keys.Keys
	DnsCache        *dns.Records
	Wss             *wss.WebSockets
	confChangeC     chan raftpb.ConfChange
	Cluster         *cluster.Cluster
	Replication     *distributed.Replication
	Kinds           *relations.RelationRegistry
	KindsRegistry   map[string]ikinds.Kind
	Manager         *manager.Manager
	VersionServer   string
}

type Kv struct {
	Value string
	Auth  string
}
