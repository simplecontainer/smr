package api

import (
	"github.com/dgraph-io/badger/v4"
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
		Config:           config,
		Keys:             &keys.Keys{},
		DnsCache:         &dns.Records{},
		Badger:           badger,
		BadgerSync:       &sync.RWMutex{},
		RelationRegistry: relations.NewDefinitionRelationRegistry(),
		Manager:          &manager.Manager{},
	}

	api.Config.Environment = startup.GetEnvironmentInfo()

	api.Manager.Config = api.Config
	api.Manager.RelationRegistry = api.RelationRegistry
	api.Manager.Keys = api.Keys
	api.Manager.DnsCache = api.DnsCache

	api.RelationRegistry.Register("containers", []string{"resource", "configuration", "certkey"})
	api.RelationRegistry.Register("gitops", []string{"certkey", "httpauth"})
	api.RelationRegistry.Register("configuration", []string{})
	api.RelationRegistry.Register("resource", []string{"configuration"})
	api.RelationRegistry.Register("certkey", []string{})
	api.RelationRegistry.Register("httpauth", []string{})

	return api
}

func (api *Api) SetupEncryptedDatabase(masterKey []byte) {
	dataKeyRotationDuration := time.Duration(3600)

	dbSecrets, err := badger.Open(badger.DefaultOptions("/home/smr-agent/smr/smr/persistent/kv-store/badger").WithEncryptionKey(masterKey).WithEncryptionKeyRotationDuration(dataKeyRotationDuration))
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	api.Badger = dbSecrets
}
