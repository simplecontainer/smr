package manager

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/objectdependency"
	"github.com/qdnqn/smr/pkg/runtime"
)
import "github.com/qdnqn/smr/pkg/config"

type Manager struct {
	Config             *config.Config
	Runtime            *runtime.Runtime
	Keys               *keys.Keys
	Badger             *badger.DB
	BadgerEncrypted    *badger.DB
	DefinitionRegistry *objectdependency.DefinitionRegistry
}
