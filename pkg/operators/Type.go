package operators

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/config"
	"smr/pkg/dns"
	"smr/pkg/manager"
	"smr/pkg/reconciler"
	"smr/pkg/registry"
	"smr/pkg/runtime"
)

// Plugin contracts
type Operator interface {
	Run(string, ...interface{}) Response
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

type Response struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             map[string]any
}
