package api

import (
	"crypto/tls"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/etcd"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/wss"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"sync"
	"time"
)

func NewApi(config *configuration.Configuration) *Api {
	api := &Api{
		User:   &authentication.User{},
		Config: config,
		Keys:   &keys.Keys{},
		DnsCache: &dns.Records{
			Lock: &sync.RWMutex{},
		},
		Wss:           wss.New(),
		Kinds:         relations.NewDefinitionRelationRegistry(),
		KindsRegistry: nil,
		Manager:       &manager.Manager{},
	}

	api.Manager.Version = api.Version
	api.Manager.User = api.User
	api.Manager.Config = api.Config
	api.Manager.Kinds = api.Kinds
	api.Manager.Keys = api.Keys
	api.Manager.DnsCache = api.DnsCache
	api.Manager.Http = clients.NewHttpClients()
	api.Manager.Wss = api.Wss

	api.Kinds.InTree()

	return api
}

func (a *Api) SetupEtcd() {
	var err error

	a.Server, err = etcd.StartEtcd(a.Config)

	if err != nil {
		panic(err)
	}

	select {
	case <-a.Server.Server.ReadyNotify():
		a.Etcd, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5 * time.Second,
		})

		if err != nil {
			panic(err)
		}
		return
	case <-time.After(60 * time.Second):
		a.Server.Server.Stop()
		panic("etcd server took too long to start!")
	}
}

func (a *Api) SetupCluster(TLSConfig *tls.Config, n *node.Node, cluster *cluster.Cluster, join bool) error {
	proposeC := make(chan string)
	insyncC := make(chan bool)
	confChangeC := make(chan raftpb.ConfChange)
	nodeUpdate := make(chan node.Node)
	nodeFinalizer := make(chan node.Node)

	getSnapshot := func() ([]byte, error) { return a.Cluster.KVStore.GetSnapshot() }

	raftNode := &raft.RaftNode{}
	rn, commitC, errorC, snapshotterReady := raft.NewRaftNode(raftNode, a.Keys, TLSConfig, n.NodeID, cluster.Cluster, join, getSnapshot, proposeC, confChangeC, nodeUpdate)

	a.Replication = distributed.New(a.Manager.Http.Clients[a.User.Username], a.User, a.Cluster.Node.NodeName, n)
	a.Replication.EventsC = make(chan KV.KV)
	a.Replication.DnsUpdatesC = a.DnsCache.Records

	a.Manager.Replication = a.Replication

	var err error
	a.Cluster.KVStore, err = raft.NewKVStore(<-snapshotterReady, proposeC, commitC, errorC, a.Replication.DataC, insyncC, join, n)

	if err != nil {
		return err
	}

	a.Cluster.RaftNode = rn
	a.Cluster.KVStore.ConfChangeC = confChangeC
	a.Cluster.KVStore.Node = n
	a.Cluster.InSync = insyncC
	a.Cluster.NodeConf = nodeUpdate
	a.Cluster.NodeFinalizer = nodeFinalizer

	a.Manager.Cluster = a.Cluster

	return nil
}
