package operators

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/config"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/runtime"
)

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) httpcontract.ResponseOperator
}

type Request struct {
	Config     *config.Config
	Runtime    *runtime.Runtime
	Registry   *registry.Registry
	Reconciler *reconciler.Reconciler
	Manager    *manager.Manager
	Badger     *badger.DB
	DnsCache   *dns.Records
	Data       map[string]any
}
