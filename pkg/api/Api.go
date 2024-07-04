package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objectdependency"
	"github.com/simplecontainer/smr/pkg/startup"
	"time"
)

func NewApi(config *configuration.Configuration, badger *badger.DB) *Api {
	api := &Api{
		Config:             config,
		Keys:               &keys.Keys{},
		Badger:             badger,
		DefinitionRegistry: objectdependency.NewDefinitionDependencyRegistry(),
		Manager:            &manager.Manager{},
	}

	api.Config.Environment = startup.GetEnvironmentInfo()

	api.Manager.Config = api.Config
	api.Manager.DefinitionRegistry = api.DefinitionRegistry

	api.DefinitionRegistry.Register("containers", []string{"resource", "configuration", "certkey"})
	api.DefinitionRegistry.Register("gitops", []string{"certkey", "httpauth"})
	api.DefinitionRegistry.Register("configuration", []string{})
	api.DefinitionRegistry.Register("resource", []string{"configuration"})
	api.DefinitionRegistry.Register("certkey", []string{})
	api.DefinitionRegistry.Register("httpauth", []string{})

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
