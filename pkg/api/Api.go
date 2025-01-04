package api

import (
	"crypto/tls"
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/startup"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"sync"
	"time"
)

func NewApi(config *configuration.Configuration, badger *badger.DB) *Api {
	api := &Api{
		User:          &authentication.User{},
		Config:        config,
		Keys:          &keys.Keys{},
		DnsCache:      &dns.Records{},
		Badger:        badger,
		BadgerSync:    &sync.RWMutex{},
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

func (api *Api) SetupEncryptedDatabase(masterKey []byte) {
	opts := badger.DefaultOptions("/home/smr-agent/smr/smr/persistent/kv-store/badger")

	opts.Dir = "/home/smr-agent/smr/smr/persistent/kv-store/badger"
	opts.ValueDir = "/home/smr-agent/smr/smr/persistent/kv-store/badger"
	opts.DetectConflicts = true
	opts.CompactL0OnClose = true
	opts.Logger = nil
	opts.SyncWrites = true
	opts.EncryptionKey = masterKey
	opts.EncryptionKeyRotationDuration = 24 * time.Hour
	opts.IndexCacheSize = 100 << 20

	dbSecrets, err := badger.Open(opts)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	api.Badger = dbSecrets
}

func (api *Api) SetupKVStore(TLSConfig *tls.Config, nodeID uint64, cluster *cluster.Cluster, join string) error {
	proposeC := make(chan string)
	confChangeC := make(chan raftpb.ConfChange)

	getSnapshot := func() ([]byte, error) { return api.Cluster.KVStore.GetSnapshot() }

	api.Cluster.RaftNode = &raft.RaftNode{}
	_, commitC, errorC, snapshotterReady := raft.NewRaftNode(api.Cluster.RaftNode, api.Keys, TLSConfig, nodeID, cluster.Cluster, join != "", getSnapshot, proposeC, confChangeC)

	var err error
	etcdC := make(chan raft.KV)
	objectC := make(chan raft.KV)

	api.Cluster.Client = api.Manager.Http
	api.Cluster.KVStore, err = raft.NewKVStore(<-snapshotterReady, api.Badger, api.Manager.Http, proposeC, commitC, errorC, etcdC, objectC)

	if err != nil {
		return err
	}

	api.Cluster.KVStore.ConfChangeC = confChangeC
	api.Cluster.KVStore.Agent = api.Config.Node

	api.Manager.Cluster = api.Cluster

	return nil
}
