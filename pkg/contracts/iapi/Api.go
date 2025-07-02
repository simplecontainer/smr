package iapi

import (
	"crypto/tls"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/version"
	"github.com/simplecontainer/smr/pkg/wss"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/embed"
)

type Api interface {
	SetupEtcd()
	SetupCluster(TLSConfig *tls.Config, n *node.Node, cluster *cluster.Cluster, join bool) error

	GetServer() *embed.Etcd
	SetServer(*embed.Etcd)

	GetEtcd() *clientv3.Client
	SetEtcd(*clientv3.Client)

	GetLeaseIdentifier() *clientv3.LeaseGrantResponse
	SetLeaseIdentifier(*clientv3.LeaseGrantResponse)

	GetUser() *authentication.User
	SetUser(*authentication.User)

	GetConfig() *configuration.Configuration
	SetConfig(*configuration.Configuration)

	GetKeys() *keys.Keys
	SetKeys(*keys.Keys)

	GetDnsCache() *dns.Records
	SetDnsCache(*dns.Records)

	GetWss() *wss.WebSockets
	SetWss(*wss.WebSockets)

	GetConfChangeC() chan raftpb.ConfChange
	SetConfChangeC(chan raftpb.ConfChange)

	GetCluster() *cluster.Cluster
	SetCluster(*cluster.Cluster)

	GetReplication() *distributed.Replication
	SetReplication(*distributed.Replication)

	GetKinds() *relations.RelationRegistry
	SetKinds(*relations.RelationRegistry)

	GetKindsRegistry() map[string]ikinds.Kind
	SetKindsRegistry(map[string]ikinds.Kind)

	GetManager() *manager.Manager
	SetManager(*manager.Manager)

	GetVersion() *version.Version
	SetVersion(*version.Version)

	HandleDns(w mdns.ResponseWriter, m *mdns.Msg)

	Kind(c *gin.Context)
	List(c *gin.Context)
	ListKind(c *gin.Context)
	GetKind(c *gin.Context)
	ProposeKind(c *gin.Context)
	CompareKind(c *gin.Context)
	SetKind(c *gin.Context)
	DeleteKind(c *gin.Context)

	ListState(c *gin.Context)
	GetState(c *gin.Context)

	ProposeKey(c *gin.Context)
	SetKey(c *gin.Context)
	RemoveKey(c *gin.Context)

	StatusCluster(c *gin.Context)
	StartCluster(c *gin.Context)
	Control(c *gin.Context)
	Nodes(c *gin.Context)
	GetNode(c *gin.Context)
	GetNodeVersion(c *gin.Context)
	AddNode(c *gin.Context)
	RemoveNode(c *gin.Context)

	Propose(c *gin.Context)
	Debug(c *gin.Context)
	Logs(c *gin.Context)
	Exec(c *gin.Context)

	CreateUser(c *gin.Context)

	Health(c *gin.Context)
	ExportClients(c *gin.Context)
	DisplayVersion(c *gin.Context)
	Events(c *gin.Context)

	MetricsHandle() gin.HandlerFunc
}
