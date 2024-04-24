package api

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/config"
	"smr/pkg/manager"
	"smr/pkg/registry"
	"smr/pkg/runtime"
)

type Api struct {
	Config   *config.Config
	Runtime  runtime.Runtime
	Registry registry.Registry
	Manager  *manager.Manager
	Badger   *badger.DB
}

type Kv struct {
	Value string
}
