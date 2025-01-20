package api

import (
	"crypto/tls"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/networking"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/startup"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"time"
)

func NewApi(config *configuration.Configuration) *Api {
	api := &Api{
		User:          &authentication.User{},
		Config:        config,
		Keys:          &keys.Keys{},
		DnsCache:      &dns.Records{},
		Kinds:         relations.NewDefinitionRelationRegistry(),
		KindsRegistry: nil,
		Manager:       &manager.Manager{},
	}

	api.Config.Environment = startup.GetEnvironmentInfo()

	api.Manager.User = api.User
	api.Manager.Config = api.Config
	api.Manager.Kinds = api.Kinds
	api.Manager.Keys = api.Keys
	api.Manager.DnsCache = api.DnsCache
	api.Manager.PluginsRegistry = []string{}
	api.Manager.Http = client.NewHttpClients()

	api.Kinds.Register("network", []string{""})
	api.Kinds.Register("containers", []string{"network", "resource", "configuration", "certkey"})
	api.Kinds.Register("container", []string{})
	api.Kinds.Register("gitops", []string{"certkey", "httpauth"})
	api.Kinds.Register("configuration", []string{})
	api.Kinds.Register("resource", []string{"configuration"})
	api.Kinds.Register("certkey", []string{})
	api.Kinds.Register("httpauth", []string{})

	return api
}

func (api *Api) SetupEtcd() {
	var err error

	api.Server, err = networking.StartEtcd(api.Config)

	if err != nil {
		panic(err)
	}

	select {
	case <-api.Server.Server.ReadyNotify():
		api.Etcd, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5 * time.Second,
		})

		if err != nil {
			panic(err)
		}
		return
	case <-time.After(60 * time.Second):
		api.Server.Server.Stop()
		panic("etcd server took too long to start!")
	}
}

func (api *Api) SetupKVStore(TLSConfig *tls.Config, nodeID uint64, cluster *cluster.Cluster, join string) error {
	proposeC := make(chan string)
	confChangeC := make(chan raftpb.ConfChange)
	nodeUpdate := make(chan node.Node)

	getSnapshot := func() ([]byte, error) { return api.Cluster.KVStore.GetSnapshot() }

	api.Cluster.RaftNode = &raft.RaftNode{}
	_, commitC, errorC, snapshotterReady := raft.NewRaftNode(api.Cluster.RaftNode, api.Keys, TLSConfig, nodeID, cluster.Cluster, join != "", getSnapshot, proposeC, confChangeC, nodeUpdate)

	api.Replication = distributed.New(api.Manager.Http.Clients[api.User.Username], api.User, api.Config.NodeName)
	api.Replication.EventsC = make(chan distributed.KV)
	api.Replication.DnsUpdatesC = api.DnsCache.Records

	api.Manager.Replication = api.Replication

	var err error
	api.Cluster.KVStore, err = raft.NewKVStore(<-snapshotterReady, api.Manager.Http, proposeC, commitC, errorC, api.Replication.DataC)

	if err != nil {
		return err
	}

	api.Cluster.KVStore.ConfChangeC = confChangeC
	api.Cluster.NodeConf = nodeUpdate
	api.Cluster.KVStore.Node = api.Config.KVStore.Node

	api.Manager.Cluster = api.Cluster

	return nil
}
