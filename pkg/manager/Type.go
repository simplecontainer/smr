package manager

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/reconciler"
	"smr/pkg/registry"
	"smr/pkg/runtime"
)
import "smr/pkg/config"

type Manager struct {
	Config     *config.Config
	Runtime    *runtime.Runtime
	Registry   *registry.Registry
	Reconciler *reconciler.Reconciler
	Badger     *badger.DB
	DnsCache   map[string]string
}
