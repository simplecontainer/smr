package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/startup"
	"sync"
	"time"
)

func NewApi(config *configuration.Configuration, badger *badger.DB) *Api {
	api := &Api{
		User:             &authentication.User{},
		Config:           config,
		Keys:             &keys.Keys{},
		DnsCache:         &dns.Records{},
		Badger:           badger,
		BadgerSync:       &sync.RWMutex{},
		RelationRegistry: relations.NewDefinitionRelationRegistry(),
		Manager:          &manager.Manager{},
	}

	api.Config.Environment = startup.GetEnvironmentInfo()

	api.Manager.User = api.User
	api.Manager.Config = api.Config
	api.Manager.RelationRegistry = api.RelationRegistry
	api.Manager.Keys = api.Keys
	api.Manager.DnsCache = api.DnsCache
	api.Manager.PluginsRegistry = []string{}
	api.Manager.Http = client.NewHttpClients()

	api.RelationRegistry.Register("network", []string{""})
	api.RelationRegistry.Register("containers", []string{"network", "resource", "configuration", "certkey"})
	api.RelationRegistry.Register("gitops", []string{"certkey", "httpauth"})
	api.RelationRegistry.Register("configuration", []string{})
	api.RelationRegistry.Register("resource", []string{"configuration"})
	api.RelationRegistry.Register("certkey", []string{})
	api.RelationRegistry.Register("httpauth", []string{})

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
