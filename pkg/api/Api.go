package api

import (
	"crypto/tls"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/etcd"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/wss"
	clientv3 "go.etcd.io/etcd/client/v3"
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
			DialTimeout: configuration.Timeout.EtcdConnectionTimeout,
		})

		if err != nil {
			panic(err)
		}
		return
	case <-time.After(configuration.Timeout.NodeStartupTimeout):
		a.Server.Server.Stop()
		panic("etcd server took too long to start!")
	}
}

func (a *Api) SetupCluster(TLSConfig *tls.Config, n *node.Node, cluster *cluster.Cluster, join bool) error {
	getSnapshot := func() ([]byte, error) { return a.Cluster.KVStore.GetSnapshot() }
	raftNode, commitC, errorC, snapshotterReady := raft.NewRaftNode(
		a.Keys,
		TLSConfig,
		n.NodeID,
		cluster.Cluster,
		join,
		a.Cluster.Replay,
		getSnapshot,
		cluster.Channels,
	)

	var err error
	a.Cluster.KVStore, err = raft.NewKVStore(a.Etcd, <-snapshotterReady, cluster.Channels, commitC, errorC, a.Replication.DataC, n, a.Cluster.Replay)

	if err != nil {
		return err
	}

	if a.Cluster.Replay {
		if len(a.Cluster.KVStore.Restore) > 0 {
			for i, v := range a.Cluster.KVStore.Restore {
				a.Replication.DataC <- *v
				a.Cluster.KVStore.Restore[i] = nil
			}
		}

		a.Cluster.KVStore.Restore = nil
		a.Cluster.Replay = false
	}

	a.Cluster.RaftNode = raftNode
	a.Cluster.KVStore.Node = n
	a.Cluster.KVStore.ConfChangeC = a.Cluster.Channels.ConfChange
	a.Cluster.InSync = a.Cluster.Channels.Insync
	a.Cluster.NodeConf = a.Cluster.Channels.NodeUpdate
	a.Cluster.NodeFinalizer = a.Cluster.Channels.NodeFinalizer

	a.Manager.Cluster = a.Cluster

	return nil
}
