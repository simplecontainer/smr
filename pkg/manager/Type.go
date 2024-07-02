package manager

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/objectdependency"
	"github.com/simplecontainer/smr/pkg/runtime"
)
import "github.com/simplecontainer/smr/pkg/config"

type Manager struct {
	Config             *config.Config
	Runtime            *runtime.Runtime
	Keys               *keys.Keys
	Badger             *badger.DB
	BadgerEncrypted    *badger.DB
	DefinitionRegistry *objectdependency.DefinitionRegistry
}
