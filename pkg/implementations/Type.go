package implementations

import (
	"smr/pkg/definitions"
	"smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Implementation(*manager.Manager, string, definitions.Definition) ([]string, []string)
}

type ImplementationInternal interface {
	ImplementationInternal(*manager.Manager, []byte) (string, error)
}
