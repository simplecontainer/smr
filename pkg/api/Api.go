package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/container"
	"github.com/simplecontainer/smr/pkg/config"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objectdependency"
	"github.com/simplecontainer/smr/pkg/registry"
	"github.com/simplecontainer/smr/pkg/runtime"
	"time"
)

func NewApi(config *config.Config, badger *badger.DB) *Api {
	api := &Api{
		Config:             config,
		Runtime:            &runtime.Runtime{},
		Registry:           &registry.Registry{},
		Keys:               &keys.Keys{},
		Badger:             badger,
		DefinitionRegistry: objectdependency.NewDefinitionDependencyRegistry(),
		Manager:            &manager.Manager{},
	}

	api.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	api.Runtime = runtime.GetRuntimeInfo()
	api.Manager.Config = api.Config
	api.Manager.Runtime = api.Runtime
	api.Manager.Keys = api.Keys
	api.Manager.Badger = badger
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

	dbSecrets, err := badger.Open(badger.DefaultOptions("/home/smr-agent/smr/smr/persistent/kv-store/badger-secrets").WithEncryptionKey(masterKey).WithEncryptionKeyRotationDuration(dataKeyRotationDuration))
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	api.BadgerEncrypted = dbSecrets
	api.Manager.BadgerEncrypted = api.BadgerEncrypted
}
